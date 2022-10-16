package http

import (
	"common/response"
	"github.com/gin-gonic/gin"
)

type MetadataController struct {
}

func NewMetadataController() *MetadataController {
	return &MetadataController{}
}

func (mc *MetadataController) Register(r gin.IRouter) {
	r.Group("metadata").
		GET("/getList", mc.Get)
}

func (mc *MetadataController) Get(c *gin.Context) {
	response.OkJson(gin.H{
		"message": "ok",
	}, c)
}
