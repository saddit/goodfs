package http

import (
	. "metaserver/internal/usecase"

	"github.com/gin-gonic/gin"
)

//TODO 版本元数据API
type VersionController struct {
	service IMetadataService
}

func NewVersionController(service IMetadataService) *VersionController {
	return &VersionController{service}
}

func (v *VersionController) Post(g *gin.Context) {

}

func (v *VersionController) Put(g *gin.Context) {

}

func (v *VersionController) Get(g *gin.Context) {

}

func (v *VersionController) Delete(g *gin.Context) {

}
