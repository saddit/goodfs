package grpc

import (
	"common/proto/pb"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"objectserver/internal/usecase/service"
	"strings"
)

type Server struct {
	*grpc.Server
}

func NewServer(service *service.MigrationService) *Server {
	serv := grpc.NewServer()
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

func (r *Server) ServeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request
		if req.ProtoMajor == 2 &&
			strings.HasPrefix(req.Header.Get("Content-Type"), "application/grpc") {
			r.ServeHTTP(c.Writer, req)
			return
		}
		c.Next()
	}
}
