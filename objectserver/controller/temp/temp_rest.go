package temp

import (
	"goodfs/lib/util"
	"goodfs/lib/util/cache"
	"goodfs/objectserver/global"
	"goodfs/objectserver/model"
	"goodfs/objectserver/service"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Patch(g *gin.Context) {
	id := g.Param("name")
	if e := service.PutFile(global.Config.TempPath, id, g.Request.Body); e != nil {
		_ = g.AbortWithError(http.StatusInternalServerError, e)
		return
	}
	g.Status(http.StatusOK)
}

func Delete(g *gin.Context) {
	id := g.Param("name")
	defer global.Cache.Delete(id)
	g.Status(http.StatusOK)
}

func Post(g *gin.Context) {
	var req model.TempPostReq
	_ = g.ShouldBindHeader(&req)
	if e := g.ShouldBindUri(&req); e != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": e.Error()})
		return
	}
	ti := &model.TempInfo{Name: req.Name, Size: req.Size}
	ti.Id = model.TempKeyPrefix + uuid.NewString()
	if !global.Cache.SetGob(ti.Id, ti) {
		g.AbortWithStatus(http.StatusServiceUnavailable)
	}
	g.Status(http.StatusOK)
	_, _ = g.Writer.Write([]byte(ti.Id))
}

func Put(g *gin.Context) {
	id := g.Param("name")
	var ti model.TempInfo
	if ok := cache.GetGob2(global.Cache, id, &ti); ok {
		if e := service.MvTmpToStorage(id, ti.Name); e != nil {
			_ = g.AbortWithError(http.StatusServiceUnavailable, e)
			return
		}
	} else {
		g.JSON(http.StatusNotFound, gin.H{"msg": "Temp file has been removed"})
		return
	}
	g.Status(http.StatusOK)
}

func HandleTempRemove(ch <-chan cache.CacheEntry) {
	log.Println("Start handle temp file removal..")
	for entry := range ch {
		if strings.HasPrefix(entry.Key, model.TempKeyPrefix) {
			var ti model.TempInfo
			if ok := util.GobDecodeGen2(entry.Value, &ti); ok {
				if e := service.DeleteFile(global.Config.TempPath, ti.Id); e != nil {
					log.Printf("Remove temp %v(name=%v) error, %v", ti.Id, ti.Name, e)
				}
			} else {
				log.Printf("Handle evicted key=%v error, value cannot cast to TempInfo", entry.Key)
			}
		}
	}
}
