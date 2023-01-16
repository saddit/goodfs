package http

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase"
	"apiserver/internal/usecase/logic"
	"apiserver/internal/usecase/webapi"
	"common/response"
	"common/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/url"
)

type MetadataController struct {
	Service usecase.IMetaService
}

func NewMetadataController(service usecase.IMetaService) *MetadataController {
	return &MetadataController{Service: service}
}

func (mc *MetadataController) Register(route gin.IRouter) {
	route.Group("metadata").
		GET("/:name/versions", mc.Versions).
		GET("/:name", mc.Get)
}

func (mc *MetadataController) Get(c *gin.Context) {
	body := struct {
		Name    string `uri:"name"`
		Bucket  string `header:"bucket" binding:"required"`
		Version int32  `form:"version"`
	}{}
	if err := entity.Bind(c, &body, false); err != nil {
		response.FailErr(err, c)
		return
	}
	data, err := mc.Service.GetMetadata(body.Name, body.Bucket, body.Version, true)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.OkJson(data, c)
}

func (mc *MetadataController) Versions(c *gin.Context) {
	body := struct {
		Page     int    `form:"page" binding:"required"`
		PageSize int    `form:"page_size" binding:"required"`
		Bucket   string `header:"bucket" binding:"required"`
		Name     string `uri:"name" binding:"min=1"`
	}{}
	if err := entity.Bind(c, &body, false); err != nil {
		response.FailErr(err, c)
		return
	}
	id := fmt.Sprint(body.Bucket, "/", body.Name)
	loc, gid, err := logic.NewHashSlot().FindMetaLocByName(id)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	loc = logic.NewDiscovery().SelectMetaByGroupID(gid, loc)
	version, total, err := webapi.ListVersion(loc, url.PathEscape(id), body.Page, body.PageSize)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	c.Header("X-Total-Count", util.ToString(total))
	if _, err = c.Writer.Write(version); err != nil {
		response.FailErr(err, c)
		return
	}
}
