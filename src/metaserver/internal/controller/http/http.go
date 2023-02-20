package http

import (
	"common/logs"
	"common/util"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"metaserver/internal/controller/grpc"
	. "metaserver/internal/usecase"
	"net/http"
)

type Server struct {
	http.Server
	grpcServer *grpc.Server
}

func NewHttpServer(port string, grpcServer *grpc.Server, service IMetadataService, bucketService BucketService) *Server {
	engine := gin.New()
	engine.Use(
		gin.LoggerWithWriter(logs.Std().Out),
		gin.RecoveryWithWriter(logs.Std().Out),
		CheckLeaderInRaftMode,
		CheckKeySlot,
	)
	engine.UseRawPath = true
	engine.UnescapePathValues = true
	//Http router
	NewMetadataController(service).RegisterRoute(engine)
	NewVersionController(service).RegisterRoute(engine)
	NewBucketController(bucketService).RegisterRoute(engine)
	return &Server{http.Server{
		Addr:    ":" + port,
		Handler: util.H2CHandler(engine, grpcServer),
	}, grpcServer}
}

func (s *Server) ListenAndServe() error {
	logs.New("http-server").Infof("server listening on %s", s.Addr)
	return s.Server.ListenAndServe()
}

func (s *Server) Shutdown(c context.Context) error {
	var err1, err2 error
	dg := util.NewWaitGroup()
	// shutdown grpc
	dg.Todo()
	go func() {
		defer dg.Done()
		err1 = s.grpcServer.Shutdown(c)
	}()
	// shutdown http
	dg.Todo()
	go func() {
		defer dg.Done()
		err2 = s.Server.Shutdown(c)
	}()
	return errors.Join(err1, err2)
}
