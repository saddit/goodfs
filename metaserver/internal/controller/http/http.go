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
	engine.Use(CheckLeaderInRaftMode)
	//Http router
	mc := NewMetadataController(service)
	engine.PUT("/metadata/:name", mc.Put)
	engine.POST("/metadata", mc.Post)
	engine.GET("/metadata/:name", mc.Get)
	engine.DELETE("/metadata/:name", mc.Delete)

	vc := NewVersionController(service)
	engine.PUT("/metadata_version/:name", vc.Put)
	engine.POST("/metadata_version/:name", vc.Post)
	engine.GET("/metadata_version/:name", vc.Get)
	engine.GET("/metadata_version/:name/list", vc.List)
	engine.DELETE("/metadata_version/:name", vc.Delete)

	return &Server{netHttp.Server{
		Addr:    addr,
		Handler: engine,
	}}
}
