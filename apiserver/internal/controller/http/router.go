package http

import (
	"apiserver/internal/controller/http/big"
	"apiserver/internal/controller/http/locate"
	"apiserver/internal/controller/http/objects"
	"apiserver/internal/controller/http/versions"
	. "apiserver/internal/usecase"
	"common/graceful"
	netHttp "net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	g      *gin.Engine
	object IObjectService
	meta   IMetaService
}

func NewHttpServer(o IObjectService, m IMetaService) *Server {
	return &Server{gin.Default(), o, m}
}

func (h *Server) ListenAndServe(addr string) {
	r := h.g.Group("/api/v1")
	versions.MetaService = h.meta
	locate.ObjectService = h.object
	objects.MetaService = h.meta
	objects.ObjectService = h.object
	big.MetaService = h.meta
	big.ObjectService = h.object

	r.PUT("/objects/:name", objects.ValidatePut, objects.Put)
	r.GET("/objects/:name", objects.Get)

	r.GET("/versions/:name", versions.Get)

	r.GET("/locate/:name", locate.Get)

	r.POST("/big/:name", big.FilterDuplicates, big.Post)
	r.HEAD("/big/:token", big.Head)
	r.PATCH("/big/:token", big.Patch)

	graceful.ListenAndServe(&netHttp.Server{Addr: addr, Handler: h.g})
}
