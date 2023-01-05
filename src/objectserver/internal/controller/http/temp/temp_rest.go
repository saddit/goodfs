package temp

import (
	"common/cache"
	"common/util"
	"log"
	"net/http"
	"objectserver/internal/entity"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/service"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
)

func Patch(g *gin.Context) {
	id := g.Param("name")
	if _, e := service.AppendFile(pool.Config.TempPath, id, g.Request.Body); e != nil {
		_ = g.AbortWithError(http.StatusInternalServerError, e)
		return
	}
	g.Status(http.StatusOK)
}

func Delete(g *gin.Context) {
	id := g.Param("name")
	defer pool.Cache.Delete(id)
	g.Status(http.StatusOK)
}

func Post(g *gin.Context) {
	var req entity.TempPostReq
	_ = g.ShouldBindHeader(&req)
	if e := g.ShouldBindUri(&req); e != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": e.Error()})
		return
	}
	tmpInfo := &entity.TempInfo{Name: req.Name, Size: req.Size}
	tmpInfo.Id = entity.TempKeyPrefix + uuid.NewString()
	if !pool.Cache.SetGob(tmpInfo.Id, tmpInfo) {
		g.AbortWithStatus(http.StatusServiceUnavailable)
	}
	g.Status(http.StatusOK)
	_, _ = g.Writer.Write(util.StrToBytes(tmpInfo.Id))
}

func Put(g *gin.Context) {
	id := g.Param("name")
	var ti *entity.TempInfo
	var ok bool
	if ti, ok = cache.GetGob[entity.TempInfo](pool.Cache, id); ok {
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

//Head 获取分片临时对象的大小
func Head(g *gin.Context) {
	s, e := os.Stat(pool.Config.TempPath + g.Param("name"))
	if e != nil {
		g.Status(http.StatusNotFound)
	} else {
		g.Header("Size", util.NumToString(s.Size()))
	}
}

//Get 获取临时对象分片
func Get(g *gin.Context) {
	req := struct {
		Name string `uri:"name" binding:"required"`
		Size int64 `form:"size" binding:"min=1"`
	}{}
	if err := entity.BindAll(g, &req, binding.Uri, binding.Query); err != nil {
		g.Status(http.StatusBadRequest)
	}
	if e := service.GetTemp(req.Name, req.Size, g.Writer); e != nil {
		log.Println(e)
		g.Status(http.StatusNotFound)
	}
}
