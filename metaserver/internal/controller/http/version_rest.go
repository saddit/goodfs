package http

import (
	"common/request"
	"common/response"
	"metaserver/internal/entity"
	. "metaserver/internal/usecase"

	"github.com/gin-gonic/gin"
)

type VersionController struct {
	service IMetadataService
}

func NewVersionController(service IMetadataService) *VersionController {
	return &VersionController{service}
}

func (v *VersionController) Post(g *gin.Context) {
	var data entity.Version
	if err := g.ShouldBindJSON(&data); err != nil {
		response.FailErr(err, g)
		return
	}
	ver, err := v.service.AddVersion(g.Param("name"), &data)
	if err != nil {
		response.FailErr(err, g)
		return
	}
	response.OkHeader(gin.H{"Version": ver}, g)
}

func (v *VersionController) Put(g *gin.Context) {
	var data entity.Version
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

func (v *VersionController) List(g *gin.Context) {
	body := struct {
		Page     int `form:"page"`
		PageSize int `form:"page_size"`
	}{}
	if err := g.ShouldBindQuery(&body); err != nil {
		response.BadRequestErr(err, g)
		return
	}
	res, err := v.service.ListVersions(g.Param("name"), body.Page, body.PageSize)
	if err != nil {
		response.FailErr(err, g)
		return
	}
	response.OkJson(res, g)
}

func (v *VersionController) Delete(g *gin.Context) {
	if ver, ok := request.GetQryInt("version", g); ok {
		if err := v.service.RemoveVersion(g.Param("name"), ver); err != nil {
			response.FailErr(err, g)
			return
		}
		response.Ok(g)
	} else {
		response.BadRequestMsg("need query param 'version'", g)
	}
}
