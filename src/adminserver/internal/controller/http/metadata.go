package http

import (
	"adminserver/internal/usecase/logic"
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
		GET("/page", mc.Page).
		GET("/versions", mc.Versions)
}

func (mc *MetadataController) Page(c *gin.Context) {
	var cond logic.MetadataCond
	if err := c.ShouldBindQuery(&c); err != nil {
		response.FailErr(err, c)
		return
	}
	res, err := logic.NewMetadata().MetadataPaging(cond)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.OkJson(res, c)
}

func (mc *MetadataController) Versions(c *gin.Context) {
	var cond logic.MetadataCond
	if err := c.ShouldBindQuery(&c); err != nil {
		response.FailErr(err, c)
		return
	}
	res, err := logic.NewMetadata().VersionPaging(cond)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	if _, err := c.Writer.Write(res); err != nil {
		response.FailErr(err, c)
		return
	}
}
