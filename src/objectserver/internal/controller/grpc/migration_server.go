package grpc

import (
	"common/graceful"
	"common/logs"
	"common/pb"
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

func (ms *MigrationServer) ReceiveData(stream pb.ObjectMigration_ReceiveDataServer) error {
	var file *os.File
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()
	for {
		data, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return stream.SendAndClose(&pb.Response{Success: false, Message: err.Error()})
		}
		if file == nil {
			if file, err = ms.Service.OpenFile(data.FileName, data.Size); err != nil {
				if os.IsExist(err) {
					break
				}
				return stream.SendAndClose(&pb.Response{Success: false, Message: err.Error()})
			}
		}
		if err = ms.Service.AppendData(file, data.Data); err != nil {
			return stream.SendAndClose(&pb.Response{Success: false, Message: err.Error()})
		}
	}
	return stream.SendAndClose(&pb.Response{Success: true})
}

func (ms *MigrationServer) FinishReceive(_ context.Context, info *pb.ObjectInfo) (*pb.Response, error) {
	if err := ms.Service.FinishObject(info); err != nil {
		return &pb.Response{Success: false, Message: err.Error()}, nil
	}
	return &pb.Response{Success: true}, nil
}

func (ms *MigrationServer) RequireSend(_ context.Context, info *pb.RequiredInfo) (*pb.Response, error) {
	logs.Std().Infof("start sending %dB data to %s", info.RequiredSize, info.TargetAddress)
	curLocate, ok := pool.Discovery.GetService(pool.Config.Registry.Name, pool.Config.Registry.ServerID, true)
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

func (ms *MigrationServer) LeaveCommand(context.Context, *pb.EmptyReq) (*pb.Response, error) {
	sizeMap, err := ms.Service.DeviationValues(false)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	curLocate, ok := pool.Discovery.GetService(pool.Config.Registry.Name, pool.Config.Registry.ServerID, true)
	if !ok {
		return nil, errors.New("could not find register address")
	}
	util.LogErr(pool.Registry.Unregister())
	if err = ms.Service.SendingTo(curLocate, sizeMap); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Response{Success: true, Message: "ok"}, nil
}

func (ms *MigrationServer) JoinCommand(context.Context, *pb.EmptyReq) (*pb.Response, error) {
	// get deviation value
	util.LogErr(pool.Registry.Register())
	curLocate, ok := pool.Discovery.GetService(pool.Config.Registry.Name, pool.Config.Registry.ServerID, true)
	if !ok {
		return nil, errors.New("could not find register address")
	}
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
			if _, err := cliMap[key].RequireSend(context.Background(), &pb.RequiredInfo{
				RequiredSize:  value,
				TargetAddress: curLocate,
			}); err != nil {
				success = false
				logs.Std().Error(err)
			}
		}(k, v)
	}
	wg.Wait()
	return &pb.Response{
		Success: success,
		Message: util.IfElse(success, "ok", "see logs for detail"),
	}, nil
}
