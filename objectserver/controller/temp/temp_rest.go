package temp

import (
	"goodfs/objectserver/config"
	"goodfs/objectserver/global"
	"goodfs/objectserver/service"
	"goodfs/util"
	"goodfs/util/cache"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PostReq struct {
	Name string `uri:"name" binding:"required"`
	Size int64  `header:"Size" binding:"required"`
}

const TempKeyPrefix = "TempInfo#"

type TempInfo struct {
	Name string
	Id   string
	Size int64
}

func Patch(g *gin.Context) {
	id := g.Param("name")
	if e := service.PutFile(config.TempPath, id, g.Request.Body); e != nil {
		g.AbortWithError(http.StatusInternalServerError, e)
		return
	}
	g.Status(http.StatusOK)
}

func Delete(g *gin.Context) {
	id := g.Param("name")
	defer global.Cache.Delete(TempKeyPrefix + id)
	g.Status(http.StatusOK)
}

func Post(g *gin.Context) {
	var req PostReq
	_ = g.ShouldBindHeader(&req)
	if e := g.ShouldBindUri(&req); e != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": e.Error()})
		return
	}
	ti := &TempInfo{Name: req.Name, Size: req.Size}
	ti.Id = uuid.NewString()
	if !global.Cache.SetGob(TempKeyPrefix+ti.Id, ti) {
		g.AbortWithStatus(http.StatusServiceUnavailable)
	}
	g.JSON(http.StatusOK, ti.Id)
}

func Put(g *gin.Context) {
	id := g.Param("name")
	if ti, ok := cache.GetGob[TempInfo](global.Cache, TempKeyPrefix+id); ok {
		if e := service.MvTmpToStorage(id, ti.Name); e != nil {
			g.AbortWithError(http.StatusServiceUnavailable, e)
			return
		}
	} else {
		g.JSON(http.StatusNotFound, gin.H{"msg": "Temp file has been removed"})
		return
	}
	g.Status(http.StatusOK)
}

func HandleTempRemove(ch <-chan cache.CacheEntry) {
	for entry := range ch {
		if strings.HasPrefix(entry.Key, TempKeyPrefix) {
			if ti, ok := util.GobDecodeGen[TempInfo](entry.Value); ok {
				if e := service.DeleteFile(config.TempPath, ti.Id); e != nil {
					log.Printf("Remove temp %v(name=%v) error, %v", ti.Id, ti.Name, e)
				}
			} else {
				log.Printf("Handle evicted key=%v error, value cannot cast to TempInfo", entry.Key)
			}
		}
	}
}
