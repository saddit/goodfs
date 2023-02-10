package http

import (
	"common/proto/msg"
	"common/request"
	"common/response"
	"github.com/gin-gonic/gin"
	"metaserver/internal/entity"
	. "metaserver/internal/usecase"
)

type VersionController struct {
	service IMetadataService
}

func NewVersionController(service IMetadataService) *VersionController {
	return &VersionController{service}
}

func (v *VersionController) RegisterRoute(engine gin.IRouter) {
	engine.PUT("/metadata_version/:name", v.Put)
	engine.POST("/metadata_version/:name", v.Post)
	engine.GET("/metadata_version/:name", v.Get)
	engine.GET("/metadata_version/:name/list", v.List)
	engine.DELETE("/metadata_version/:name", v.Delete)
	engine.PATCH("/version/locate", v.UpdateLocates)
	engine.GET("/version/list", v.ListByCond)
}

func (v *VersionController) Post(g *gin.Context) {
	var data msg.Version
	if err := g.ShouldBindJSON(&data); err != nil {
		response.FailErr(err, g)
		return
	}
	data.Sequence = 0
	ver, err := v.service.AddVersion(g.Param("name"), &data)
	if err != nil {
		response.FailErr(err, g)
		return
	}
	response.CreatedHeader(gin.H{"Version": ver}, g)
}

func (v *VersionController) Put(g *gin.Context) {
	var data msg.Version
	if err := g.ShouldBindJSON(&data); err != nil {
		response.FailErr(err, g)
		return
	}
	if s, ok := request.GetQryInt("version", g); ok {
		err := v.service.UpdateVersion(g.Param("name"), s, &data)
		if err != nil {
			response.FailErr(err, g)
			return
		}
		response.Ok(g)
	} else {
		response.BadRequestMsg("need query param 'version'", g)
	}
}

func (v *VersionController) Get(g *gin.Context) {
	if ver, ok := request.GetQryInt("version", g); ok {
		data, err := v.service.GetVersion(g.Param("name"), ver)
		if err != nil {
			response.FailErr(err, g)
			return
		}
		response.OkJson(data, g)
	} else {
		response.BadRequestMsg("need query param 'version'", g)
	}
}

func (v *VersionController) ListByCond(c *gin.Context) {
	body := struct {
		Hash string `form:"hash" binding:"required"`
	}{}
	if err := c.ShouldBindQuery(&body); err != nil {
		response.BadRequestErr(err, c)
		return
	}
	res, err := v.service.FindByHash(body.Hash)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.OkJson(res, c)
}

func (v *VersionController) List(g *gin.Context) {
	body := struct {
		Page     int `form:"page" binding:"required"`
		PageSize int `form:"page_size" binding:"required"`
	}{}
	if err := g.ShouldBindQuery(&body); err != nil {
		response.BadRequestErr(err, g)
		return
	}
	res, total, err := v.service.ListVersions(g.Param("name"), body.Page, body.PageSize)
	if err != nil {
		response.FailErr(err, g)
		return
	}
	response.Exec(g).
		Header(gin.H{"X-Total-Count": total}).
		JSON(res)
}

func (v *VersionController) Delete(g *gin.Context) {
	if ver, ok := request.GetQryInt("version", g); ok {
		if err := v.service.RemoveVersion(g.Param("name"), ver); err != nil {
			response.FailErr(err, g)
			return
		}
		response.NoContent(g)
	} else {
		response.BadRequestMsg("need query param 'version'", g)
	}
}

func (v *VersionController) UpdateLocates(c *gin.Context) {
	body := struct {
		Locate      string `json:"locate" binding:"required"`
		Hash        string `json:"hash" binding:"required"`
		LocateIndex int    `json:"locateIndex"`
	}{}
	if err := entity.BindAll(c, &body, entity.FullBindings...); err != nil {
		response.BadRequestErr(err, c)
		return
	}
	if err := v.service.UpdateLocates(body.Hash, body.LocateIndex, body.Locate); err != nil {
		response.FailErr(err, c)
		return
	}
	response.Ok(c)
}
