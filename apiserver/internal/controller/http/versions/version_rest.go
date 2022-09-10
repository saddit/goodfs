package versions

import (
	"apiserver/internal/entity"
	"common/response"
	"github.com/gin-gonic/gin"
)

func Get(g *gin.Context) {
	name := g.Param("name")
	//TODO 分页
	if meta, ok := MetaService.GetMetadata(name, int32(entity.VerModeALL)); ok {
		response.OkJson(meta, g)
	} else {
		response.NotFound(g).Abort()
	}
}
