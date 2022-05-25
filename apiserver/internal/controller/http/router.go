package http

import (
	"apiserver/internal/controller/http/big"
	"apiserver/internal/controller/http/locate"
	"apiserver/internal/controller/http/objects"
	"apiserver/internal/controller/http/versions"
	. "apiserver/internal/usecase"

	"github.com/gin-gonic/gin"
)

func RegisterRouter(r gin.IRouter, o IObjectService, m IMetaService) {

	versions.MetaService = m
	locate.ObjectService = o
	objects.MetaService = m
	objects.ObjectService = o
	big.MetaService = m
	big.ObjectService = o

	r.PUT("/objects/:name", objects.ValidatePut, objects.Put)
	r.GET("/objects/:name", objects.Get)

	r.GET("/versions/:name", versions.Get)

	r.GET("/locate/:name", locate.Get)

	r.POST("/big/:name", big.FilterDuplicates, big.Post)
	r.HEAD("/big/:token", big.Head)
	r.PATCH("/big/:token", big.Patch)
}
