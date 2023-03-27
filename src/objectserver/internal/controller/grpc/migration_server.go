package grpc

import (
	"common/graceful"
	"common/logs"
	"common/proto/pb"
	"common/util"
	"context"
	"errors"
	"io"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/service"
	"os"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MigrationServer struct {
	pb.UnimplementedObjectMigrationServer
	Service *service.MigrationService
}

func NewMigrationServer(service *service.MigrationService) *MigrationServer {
	return &MigrationServer{Service: service}
}

func (ms *MigrationServer) ReceiveData(stream pb.ObjectMigration_ReceiveDataServer) (err error) {
	var file io.WriteCloser
	var data *pb.ObjectData
	defer util.CloseAndLog(file)
	logs.Std().Debug("start receive data...")
	for {
		data, err = stream.Recv()
		if err == io.EOF {
			return stream.SendMsg(&pb.Response{Success: true})
		}
		if err != nil {
			break
		}
		logs.Std().Debugf("received filename chunk %s", data.FileName)
		if file == nil {
			if data.FileName == "" {
				err = errors.New("received FileName should not be empty")
				break
			}
			if file, err = ms.Service.OpenFile(data.FileName, data.Size); err != nil {
				if os.IsExist(err) {
					logs.Std().Debugf("receive duplicate data %s success, close but send success result", data.FileName)
					return stream.SendAndClose(&pb.Response{Success: true})
				}
				break
			}
		}
		if _, err = file.Write(data.Data); err != nil {
			break
		}
	}
	logs.Std().Errorf("receive data err: %s", err.Error())
	return stream.SendAndClose(&pb.Response{Success: false, Message: err.Error()})
}

func (ms *MigrationServer) FinishReceive(_ context.Context, info *pb.ObjectInfo) (*pb.Response, error) {
	if err := ms.Service.FinishObject(info); err != nil {
		return &pb.Response{Success: false, Message: err.Error()}, nil
	}
	return &pb.Response{Success: true}, nil
}

func (ms *MigrationServer) RequireSend(_ context.Context, info *pb.RequiredInfo) (*pb.Response, error) {
	logs.Std().Infof("start sending %dB data to %s", info.RequiredSize, info.TargetAddress)
	curLocate, ok := pool.Discovery.GetService(pool.Config.Registry.Name, pool.Config.Registry.SID())
	if !ok {
		return nil, errors.New("could not find register address")
	}
	if err := ms.Service.SendingTo(curLocate, map[string]int64{
		info.TargetAddress: info.RequiredSize,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Response{Success: true, Message: "ok"}, nil
}

func (ms *MigrationServer) LeaveCommand(c context.Context, _ *pb.EmptyReq) (*pb.Response, error) {
	curLocate, ok := pool.Discovery.GetService(pool.Config.Registry.Name, pool.Config.Registry.SID())
	if !ok {
		return &pb.Response{Success: false, Message: "it's an unregister server"}, nil
	}
	sizeMap, err := ms.Service.DeviationValues(false)
	if err != nil {
		return &pb.Response{Success: false, Message: err.Error()}, nil
	}
	if err = pool.CloseGraceful(); err != nil {
		return &pb.Response{Success: false, Message: "close pool fail: " + err.Error()}, nil
	}
	if err = ms.Service.SendingTo(curLocate, sizeMap); err != nil {
		return &pb.Response{Success: false, Message: err.Error()}, nil
	}
	return &pb.Response{Success: true, Message: "ok"}, nil
}

func (ms *MigrationServer) JoinCommand(context.Context, *pb.EmptyReq) (*pb.Response, error) {
	if err := pool.OpenGraceful(); err != nil {
		return &pb.Response{Success: false, Message: "open pool fail"}, nil
	}
	curLocate, ok := pool.Discovery.GetService(pool.Config.Registry.Name, pool.Config.Registry.SID())
	if !ok {
		return &pb.Response{Success: false, Message: "could not find register address of server"}, nil
	}
	// get deviation value
	sizeMap, err := ms.Service.DeviationValues(true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// dail connection to all servers
	cliMap := make(map[string]pb.ObjectMigrationClient, len(sizeMap))
	conns := make([]*grpc.ClientConn, 0, len(sizeMap))
	defer func() {
		for _, cc := range conns {
			util.LogErr(cc.Close())
		}
	}()
	for k := range sizeMap {
		cc, err := grpc.Dial(k, grpc.WithInsecure())
		if err != nil {
			return nil, status.Error(codes.Unavailable, err.Error())
		}
		conns = append(conns, cc)
		cliMap[k] = pb.NewObjectMigrationClient(cc)
	}
	wg := sync.WaitGroup{}
	success := true
	// sending request to all servers
	for k, v := range sizeMap {
		wg.Add(1)
		go func(key string, value int64) {
			defer graceful.Recover()
			defer wg.Done()
			if _, inner := cliMap[key].RequireSend(context.Background(), &pb.RequiredInfo{
				RequiredSize:  value,
				TargetAddress: curLocate,
			}); inner != nil {
				success = false
				logs.Std().Error(inner)
			}
		}(k, v)
	}
	wg.Wait()
	return &pb.Response{
		Success: success,
		Message: util.IfElse(success, "ok", "see logs for detail"),
	}, nil
}
