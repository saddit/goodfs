package grpc

import (
	"common/pb"
	"common/util"
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"objectserver/internal/usecase/service"
)

type Server struct {
	*grpc.Server
	Port string
}

func NewRpcServer(port string, service *service.MigrationService) *Server {
	serv := grpc.NewServer()
	pb.RegisterObjectMigrationServer(serv, NewMigrationServer(service))
	return &Server{serv, port}
}

func (r *Server) ListenAndServe() error {
	if r.Server == nil {
		return nil
	}
	sock, err := net.Listen("tcp", util.GetHostPort(r.Port))
	if err != nil {
		panic(err)
	}
	log.Infof("rpc server listening on %s", sock.Addr().String())
	return r.Serve(sock)
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
