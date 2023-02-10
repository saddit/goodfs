package http

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase"
	"apiserver/internal/usecase/grpcapi"
	"apiserver/internal/usecase/logic"
	"common/response"
	"common/util"
	"fmt"
	"github.com/gin-gonic/gin"
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
	serverId, err := logic.NewHashSlot().KeySlotLocation(id)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	ip, err := logic.NewDiscovery().SelectMetaServerGRPC(serverId)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	version, total, err := grpcapi.ListVersion(ip, id, body.Page, body.PageSize)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	c.Header("X-Total-Count", util.ToString(total))
	response.OkJson(version, c)
}
