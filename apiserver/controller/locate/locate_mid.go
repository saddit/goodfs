package locate

import (
	"github.com/gin-gonic/gin"
	"goodfs/apiserver/global"
	"goodfs/apiserver/model"
	"goodfs/apiserver/service"
	"net/http"
)

func FilterExisting(g *gin.Context) {
	name := g.Param("name")
	if !global.ExistFilter.Lookup([]byte(name)) {
		g.AbortWithStatus(http.StatusNotFound)
	}
}

func ChangeExisting(g *gin.Context) {
	key := []byte(g.Param("name"))
	if g.Request.Response.StatusCode == http.StatusOK {
		_ = service.SendExistingSyncMsg(key, model.SyncInsert)
	} else {
		_ = service.SendExistingSyncMsg(key, model.SyncDelete)
	}
}
