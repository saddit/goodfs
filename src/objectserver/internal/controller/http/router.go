package http

import (
	"common/logs"
	"net/http"
	"objectserver/internal/controller/http/objects"
	"objectserver/internal/controller/http/stat"
	"objectserver/internal/controller/http/temp"

	"github.com/gin-gonic/gin"
)

type Server struct {
	*http.Server
}

func NewHttpServer(port string) *Server {
	r := gin.New()
	r.UseH2C = true
	r.Use(gin.LoggerWithWriter(logs.Std().Out), gin.RecoveryWithWriter(logs.Std().Out))
	r.GET("/objects/:name", objects.GetFromCache, objects.Get)
	r.HEAD("/objects/:name", objects.Head)
	r.PUT("/objects/:name", temp.FilterEmptyRequest, objects.Put)
	r.DELETE("/objects/:name", objects.Delete)

	r.POST("/temp/:name", temp.Post)
	r.PATCH("/temp/:name", temp.FilterExpired, temp.FilterEmptyRequest, temp.Patch)
	r.DELETE("/temp/:name", temp.FilterExpired, temp.Delete)
	r.HEAD("/temp/:name", temp.FilterExpired, temp.Head)
	r.GET("/temp/:name", temp.FilterExpired, temp.Get)
	r.PUT("/temp/:name", temp.FilterExpired, temp.Put)

	r.GET("/ping", stat.Ping)
	r.GET("/stat", stat.Info)

	return &Server{&http.Server{Addr: ":" + port, Handler: r.Handler()}}
}

func (h *Server) ListenAndServe() error {
	logs.Std().Infof("http server listen on: %s", h.Addr)
	return h.Server.ListenAndServe()
}
