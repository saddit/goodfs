package objects

import (
	"common/logs"
	"net/http"
	"objectserver/internal/usecase/pool"

	"github.com/gin-gonic/gin"
)

func GetFromCache(g *gin.Context) {
	name := g.Param("name")
	if bt, ok := pool.Cache.HasGet(name); ok {
		if _, e := g.Writer.Write(bt); e != nil {
			logs.Std().Debugf("match file cache %v, but written to response error: %v", name, e)
			g.AbortWithStatus(http.StatusInternalServerError)
		} else {
			logs.Std().Debugf("match file cache %v", name)
			g.AbortWithStatus(http.StatusOK)
		}
	}
}
