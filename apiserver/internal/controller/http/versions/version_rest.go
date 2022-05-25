package versions

import (
	"apiserver/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Get(g *gin.Context) {
	name := g.Param("name")

	if meta, ok := MetaService.GetMetadata(name, int32(entity.VerModeALL)); ok {
		g.JSON(http.StatusOK, meta)
	} else {
		g.AbortWithStatus(http.StatusNotFound)
	}
}
