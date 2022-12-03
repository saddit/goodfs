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
		GET("/versions", mc.Versions).
		POST("/migration", mc.Migration).
		GET("/slots_detail", mc.SlotsDetail)
}

func (mc *MetadataController) Page(c *gin.Context) {
	var cond logic.MetadataCond
	if err := c.ShouldBindQuery(&cond); err != nil {
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
	if err := c.ShouldBindQuery(&cond); err != nil {
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

func (mc *MetadataController) Migration(c *gin.Context) {
	body := struct {
		SrcServerId  string   `json:"srcServerId" binding:"required"`
		DestServerId string   `json:"destServerId" binding:"required"`
		Slots        []string `json:"slots" binding:"required"`
	}{}
	if err := logic.NewMetadata().StartMigration(body.SrcServerId, body.DestServerId, body.Slots); err != nil {
		response.FailErr(err, c)
		return
	}
	response.Ok(c)
}

func (mc *MetadataController) SlotsDetail(c *gin.Context) {
	detail, err := logic.NewMetadata().GetSlotsDetail()
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.OkJson(detail, c)
}
