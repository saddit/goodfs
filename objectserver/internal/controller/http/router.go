package http

import (
	"common/graceful"
	"net/http"
	"objectserver/internal/controller/http/objects"
	"objectserver/internal/controller/http/temp"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	engine *gin.Engine
}

func NewHttpServer() *HttpServer {
	return &HttpServer{gin.Default()}
}

func (h *HttpServer) ListenAndServe(addr string) {
	r := h.engine
	r.GET("/objects/:name", objects.GetFromCache, objects.Get, objects.SaveToCache)
	//Deprecated
	r.PUT("/objects/:name", objects.SaveToCache, objects.Put, objects.RemoveCache)
	r.DELETE("/objects/:name", objects.Delete, objects.RemoveCache)

	r.POST("/temp/:name", temp.Post)
	r.PATCH("/temp/:name", temp.FilterExpired, temp.Patch)
	r.DELETE("/temp/:name", temp.FilterExpired, temp.Delete)
	r.HEAD("/temp/:name", temp.FilterExpired, temp.Head)
	r.GET("/temp/:name", temp.FilterExpired, temp.Get)
	r.PUT("/temp/:name", temp.FilterExpired, temp.Put)

	graceful.ListenAndServe(&http.Server{Addr: addr, Handler: h.engine})
}
