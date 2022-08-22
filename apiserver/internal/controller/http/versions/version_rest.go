package versions

import (
	"apiserver/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Get(g *gin.Context) {
	name := g.Param("name")
	//TODO 分页
	if meta, ok := MetaService.GetMetadata(name, int32(entity.VerModeALL)); ok {
		g.JSON(http.StatusOK, meta)
	} else {
		g.AbortWithStatus(http.StatusNotFound)
	}
}
