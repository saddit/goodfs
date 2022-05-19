package controller

import (
	"goodfs/apiserver/controller/big"
	"goodfs/apiserver/controller/locate"
	"goodfs/apiserver/controller/objects"
	"goodfs/apiserver/controller/versions"

	"github.com/gin-gonic/gin"
)

func Router(r gin.IRouter) {
	r.PUT("/objects/:name", objects.ValidatePut, objects.Put)
	r.GET("/objects/:name", objects.Get)

	r.GET("/versions/:name", versions.Get)

	r.GET("/locate/:name", locate.Get)

	r.POST("/big/:name", big.FilterDuplicates, big.Post)
	r.HEAD("/big/:token", big.Head)
	r.PATCH("/big/:token", big.Patch)
}
