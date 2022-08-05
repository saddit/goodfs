package http

import (
	. "metaserver/internal/usecase"

	"github.com/gin-gonic/gin"
)

//TODO 元数据API
type MetadataController struct {
	service IMetadataService
}

func NewMetadataController(service IMetadataService) *MetadataController {
	return &MetadataController{service}
}

func (m *MetadataController) Post(g *gin.Context) {

}

func (m *MetadataController) Put(g *gin.Context) {

}

func (m *MetadataController) Get(g *gin.Context) {

}

func (m *MetadataController) Delete(g *gin.Context) {

}
