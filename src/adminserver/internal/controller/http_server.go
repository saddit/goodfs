package controller

import (
	http2 "adminserver/internal/controller/http"
	"common/logs"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

type HttpServer struct {
	http.Server
}

func NewHttpServer(addr string, webFs static.ServeFileSystem) *HttpServer {
	eng := gin.Default()
	randSec := uuid.New()
	eng.Use(static.Serve("/", webFs))
	eng.Use(sessions.Sessions("dfs-admin", cookie.NewStore(randSec[:])))

	route := eng.Group("/api")
	http2.NewMetadataController().Register(route)
	http2.NewServerStateController().Register(route)
	http2.NewObjectsController().Register(route)

	return &HttpServer{Server: http.Server{Handler: eng, Addr: addr}}
}

func (s *HttpServer) ListenAndServe() error {
	logs.New("http-server").Infof("server listening on %s", s.Addr)
	return s.Server.ListenAndServe()
}
