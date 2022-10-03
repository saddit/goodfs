package grpc

import (
	"common/pb"
	"context"
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
	//TODO(feat): join command
	return &pb.Response{Success: true, Message: "ok"}, nil
}
