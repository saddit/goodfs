package http

import (
	"common/logs"
	"common/util"
	"net/http"
	"objectserver/internal/controller/http/objects"
	"objectserver/internal/controller/http/temp"
	"objectserver/internal/db"

	"github.com/gin-gonic/gin"
)

type Server struct {
	*http.Server
}

func NewHttpServer(addr string, capDB *db.ObjectCapacity) *Server {
	r := gin.New()
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

	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/stat", func(c *gin.Context) { c.Header("Capacity", util.ToString(capDB.Capacity())) })
	return &Server{&http.Server{Addr: addr, Handler: r}}
}

func (h *Server) ListenAndServe() error {
	logs.Std().Infof("http server listen on: %s", h.Addr)
	return h.Server.ListenAndServe()
}
