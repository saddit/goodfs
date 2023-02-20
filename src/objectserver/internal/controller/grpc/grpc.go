package grpc

import (
	"common/datasize"
	"common/proto/pb"
	"common/util"
	"context"
	"errors"
	"google.golang.org/grpc"
	"objectserver/internal/usecase/service"
)

type Server struct {
	*grpc.Server
}

func NewServer(service *service.MigrationService) *Server {
	serv := grpc.NewServer(
		grpc.MaxConcurrentStreams(100),
		grpc.MaxRecvMsgSize(int(8*datasize.MB)),
		util.CommonUnaryInterceptors(),
		util.CommonStreamInterceptors(),
	)

	pb.RegisterObjectMigrationServer(serv, NewMigrationServer(service))
	pb.RegisterConfigServiceServer(serv, &ConfigServiceServer{})
	return &Server{serv}
}

func (r *Server) Shutdown(ctx context.Context) error {
	if r.Server == nil {
		return nil
	}
	finish := make(chan struct{})
	go func() {
		defer close(finish)
		r.Server.GracefulStop()
	}()
	select {
	case <-ctx.Done():
		r.Server.Stop()
		return errors.New("graceful stop grpc server timeout")
	case <-finish:
		return nil
	}
}
