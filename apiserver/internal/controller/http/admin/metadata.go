package admin

import "github.com/gin-gonic/gin"

type MetadataController struct {
}

func NewMetadataController() *MetadataController {
	return &MetadataController{}
}

func (mc *MetadataController) Register(r gin.IRoutes) {

}
