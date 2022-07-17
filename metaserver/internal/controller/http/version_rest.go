package http

import (
	. "metaserver/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
)

//TODO 版本元数据API 操作成功后Apply到Raft自动机中
type VersionController struct {
	raft    *raft.Raft			//raft maybe nil which means disabled raft
	service IMetadataService
}

func NewVersionController(rf *raft.Raft, service IMetadataService) *VersionController {
	return &VersionController{rf, service}
}

func (v *VersionController) Post(g *gin.Context) {

}

func (v *VersionController) Put(g *gin.Context) {

}

func (v *VersionController) Get(g *gin.Context) {

}

func (v *VersionController) Delete(g *gin.Context) {

}
