package grpc

import (
	"common/graceful"
	"common/logs"
	"common/pb"
	"common/util"
	"context"
	"io"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/service"
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

func (ms *MigrationServer) ReceiveObject(stream pb.ObjectMigration_ReceiveObjectServer) error {
	for {
		data, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err = ms.Service.Received(data); err != nil {
			util.LogErr(stream.Send(&pb.Response{Success: false, Message: err.Error()}))
		} else {
			util.LogErr(stream.Send(&pb.Response{Success: true}))
		}
	}
	return nil
}

func (ms *MigrationServer) RequireSend(_ context.Context, info *pb.RequiredInfo) (*pb.Response, error) {
	logs.Std().Infof("start sending %dB data to %s", info.RequiredSize, info.TargetAddress)
	if err := ms.Service.SendingTo(map[string]int64{
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
	if err = ms.Service.SendingTo(sizeMap); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Response{Success: true, Message: "ok"}, nil
}

func (ms *MigrationServer) JoinCommand(context.Context, *pb.EmptyReq) (*pb.Response, error) {
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
			if _, err := cliMap[key].RequireSend(context.Background(), &pb.RequiredInfo{
				RequiredSize:  value,
				TargetAddress: util.GetHostPort(pool.Config.RpcPort),
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
