package http

import (
	. "metaserver/internal/usecase"
	netHttp "net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	netHttp.Server
}

func NewHttpServer(addr string, service IMetadataService) *Server {
	engine := gin.Default()
	engine.Use(CheckLeaderInRaftMode, CheckKeySlot)
	//Http router
	mc := NewMetadataController(service)
	mc.RegisterRoute(engine)

	vc := NewVersionController(service)
	vc.RegisterRoute(engine)

	return &Server{netHttp.Server{
		Addr:    addr,
		Handler: engine,
	}}
}
