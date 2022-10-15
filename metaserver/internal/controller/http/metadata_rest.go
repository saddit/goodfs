package http

import (
	"common/response"
	"github.com/gin-gonic/gin"
	"metaserver/internal/entity"
	. "metaserver/internal/usecase"
	"metaserver/internal/usecase/logic"
	"net/http"
)

type MetadataController struct {
	service IMetadataService
}

func NewMetadataController(service IMetadataService) *MetadataController {
	return &MetadataController{service}
}

func (m *MetadataController) RegisterRoute(engine gin.IRouter) {
	engine.PUT("/metadata/:name", m.Put)
	engine.POST("/metadata", m.Post)
	engine.GET("/metadata/:name", m.Get)
	engine.DELETE("/metadata/:name", m.Delete)
	engine.GET("/metadata/list", m.List)
}

func (m *MetadataController) Post(g *gin.Context) {
	var data entity.Metadata
	if err := g.ShouldBindJSON(&data); err != nil {
		response.FailErr(err, g)
		return
	}
	if ok, other := logic.NewHashSlot().IsKeyOnThisServer(data.Name); !ok {
		response.Exec(g).Redirect(http.StatusSeeOther, other)
		return
	}
	if err := m.service.AddMetadata(&data); err != nil {
		response.FailErr(err, g)
		return
	}
	response.Created(g)
}

func (m *MetadataController) Put(g *gin.Context) {
	var data entity.Metadata
	_ = g.ShouldBindJSON(&data)
	if err := m.service.UpdateMetadata(g.Param("name"), &data); err != nil {
		response.FailErr(err, g)
		return
	}
	response.Ok(g)
}

func (m *MetadataController) Get(g *gin.Context) {
	qry := struct {
		Version int `form:"version"`
	}{}
	if err := g.ShouldBindQuery(&qry); err != nil {
		response.FailErr(err, g)
		return
	}
	meta, vers, err := m.service.GetMetadata(g.Param("name"), qry.Version)
	if err != nil {
		response.FailErr(err, g)
		return
	}
	var versionList []*entity.Version
	if vers != nil {
		versionList = append(versionList, vers)
	}
	// metadata and version format
	response.OkJson(struct {
		*entity.Metadata
		Versions []*entity.Version `json:"versions,omitempty"`
	}{meta, versionList}, g)
}

func (m *MetadataController) Delete(g *gin.Context) {
	err := m.service.RemoveMetadata(g.Param("name"))
	if err != nil {
		response.FailErr(err, g)
		return
	}
	response.NoContent(g)
}

func (m *MetadataController) List(c *gin.Context) {
	req := struct {
		Prefix   string `form:"prefix"`
		PageSize int    `form:"page_size" binding:"required,lte=10000"`
	}{}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailErr(err, c)
		return
	}
	res, err := m.service.ListMetadata(req.Prefix, req.PageSize)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.OkJson(res, c)
}
