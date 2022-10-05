package grpc

import (
	"common/pb"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"objectserver/internal/usecase/service"
)

type MigrationServer struct {
	pb.UnimplementedObjectMigrationServer
	Service *service.MigrationService
}

func NewMigrationServer(service *service.MigrationService) *MigrationServer {
	return &MigrationServer{Service: service}
}

func (ms *MigrationServer) ReceiveObject(stream pb.ObjectMigration_ReceiveObjectServer) error {
	//TODO(feat): receive api
	return nil
}

func (ms *MigrationServer) RequireSend(context.Context, *pb.RequiredInfo) (*pb.Response, error) {
	//TODO(feat): require sending api
	return &pb.Response{Success: true, Message: "ok"}, nil
}

func (ms *MigrationServer) LeaveCommand(context.Context, *pb.EmptyReq) (*pb.Response, error) {
	//TODO(feat): leave command
	return &pb.Response{Success: true, Message: "ok"}, nil
}

func (ms *MigrationServer) JoinCommand(context.Context, *pb.EmptyReq) (*pb.Response, error) {
	sizeMap, err := ms.Service.DeviationValues(true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	_ = ms.Service.SendingTo(sizeMap)
	return &pb.Response{Success: true, Message: "ok"}, nil
}
