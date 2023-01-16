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

func NewHttpServer(addr string, service IMetadataService) *Server {
	engine := gin.New()
	engine.Use(
		gin.LoggerWithWriter(logs.Std().Out), 
		gin.RecoveryWithWriter(logs.Std().Out), 
		CheckLeaderInRaftMode, 
		CheckKeySlot,
	)
	//Http router
	NewMetadataController(service).RegisterRoute(engine)
	NewVersionController(service).RegisterRoute(engine)

	return &Server{netHttp.Server{
		Addr:    addr,
		Handler: engine,
	}}
}

func (s *Server) ListenAndServe() error {
	logs.New("http-server").Infof("server listening on %s", s.Addr)
	return s.Server.ListenAndServe()
}
