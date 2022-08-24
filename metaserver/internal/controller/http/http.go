package http

import (
	. "metaserver/internal/usecase"
	netHttp "net/http"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	netHttp.Server
}

func NewHttpServer(addr string, service IMetadataService) *HttpServer {
	engine := gin.Default()
	//Http router
	mc := NewMetadataController(service)
	engine.PUT("/metadata/:name", mc.Put)
	engine.POST("/metadata/:name", mc.Post)
	engine.GET("/metadata/:name", mc.Get)
	engine.DELETE("/metadata/:name", mc.Delete)

	vc := NewVersionController(service)
	engine.PUT("/metadata_version/:name", vc.Put)
	engine.POST("/metadata_version/:name", vc.Post)
	engine.GET("/metadata_version/:name", vc.Get)
	engine.DELETE("/metadata_version/:name", vc.Delete)

	return &HttpServer{netHttp.Server{
		Addr:    addr,
		Handler: engine,
	}}
}
