package controller

import (
	http2 "adminserver/internal/controller/http"
	"common/logs"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
)

type HttpServer struct {
	http.Server
}

func NewHttpServer(addr string, resourcePath string) *HttpServer {
	eng := gin.Default()

	eng.Use(static.Serve("/", static.LocalFile(filepath.Join(resourcePath, "web"), false)))

	route := eng.Group("/api")
	http2.NewMetadataController().Register(route)
	http2.NewServerStateController().Register(route)

	return &HttpServer{Server: http.Server{Handler: eng, Addr: addr}}
}

func (s *HttpServer) ListenAndServe() error {
	logs.New("http-server").Infof("server listening on %s", s.Addr)
	return s.Server.ListenAndServe()
}
