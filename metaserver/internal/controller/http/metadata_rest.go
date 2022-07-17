package http

import (
	. "metaserver/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
)

//TODO 元数据API 操作成功后Apply到Raft自动机中
type MetadataController struct {
	raft    *raft.Raft			//raft maybe nil which means disabled raft
	service IMetadataService
}

func NewMetadataController(rf *raft.Raft, service IMetadataService) *MetadataController {
	return &MetadataController{rf, service}
}

func (m *MetadataController) Post(g *gin.Context) {

}

func (m *MetadataController) Put(g *gin.Context) {

}

func (m *MetadataController) Get(g *gin.Context) {

}

func (m *MetadataController) Delete(g *gin.Context) {

}
