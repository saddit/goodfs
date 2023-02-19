package http

import (
	"common/logs"
	. "metaserver/internal/usecase"
	netHttp "net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	netHttp.Server
}

func NewHttpServer(addr string, service IMetadataService, bucketService BucketService) *Server {
	engine := gin.New()
	engine.Use(
		gin.LoggerWithWriter(logs.Std().Out),
		gin.RecoveryWithWriter(logs.Std().Out),
		CheckLeaderInRaftMode,
		CheckKeySlot,
	)
	engine.UseH2C = true
	engine.UseRawPath = true
	engine.UnescapePathValues = true
	//Http router
	NewMetadataController(service).RegisterRoute(engine)
	NewVersionController(service).RegisterRoute(engine)
	NewBucketController(bucketService).RegisterRoute(engine)
	return &Server{netHttp.Server{
		Addr:    addr,
		Handler: engine.Handler(),
	}}
}

func (s *Server) ListenAndServe() error {
	logs.New("http-server").Infof("server listening on %s", s.Addr)
	return s.Server.ListenAndServe()
}
