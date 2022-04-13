package versions

import (
	"goodfs/apiserver/repository/metadata"
	"goodfs/apiserver/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Get(g *gin.Context) {
	name := g.Param("name")

	if meta, ok := service.GetMetaData(name, metadata.VerModeALL); ok {
		g.JSON(http.StatusOK, meta)
	} else {
		g.AbortWithStatus(http.StatusNotFound)
	}
}
