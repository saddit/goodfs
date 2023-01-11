package temp

import (
	"common/cache"
	"common/util"
	xmath "common/util/math"
	"net/http"
	"objectserver/internal/entity"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/service"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
)

func Patch(g *gin.Context) {
	id := g.Param("name")
	fullPath := filepath.Join(pool.Config.TempPath, id)
	// only allow last chunck can not be power of 4KB
	if _, err := service.WriteFile(fullPath, g.Request.Body); err != nil {
		util.LogErr(g.AbortWithError(http.StatusInternalServerError, err))
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
	if err := entity.BindAll(g, &req, binding.Header, binding.Uri); err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}
	tmpInfo := &entity.TempInfo{
		Name: req.Name, 
		Size: req.Size,
		Id: entity.TempKeyPrefix + uuid.NewString(),
	}
	if !pool.Cache.SetGob(tmpInfo.Id, tmpInfo) {
		g.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	g.Status(http.StatusOK)
	_, _ = g.Writer.Write(util.StrToBytes(tmpInfo.Id))
}

func Put(g *gin.Context) {
	id := g.Param("name")
	var ti *entity.TempInfo
	var ok bool
	if ti, ok = cache.GetGob[entity.TempInfo](pool.Cache, id); ok {
		if err := service.MvTmpToStorage(id, ti.Name); err != nil {
			status := util.IfElse(os.IsNotExist(err), http.StatusNotFound, http.StatusInternalServerError)
			util.LogErr(g.AbortWithError(status, err))
			return
		}
	} else {
		g.JSON(http.StatusNotFound, gin.H{"msg": "Temp file has been removed"})
		return
	}
	pool.Cache.Delete(id)
	g.Status(http.StatusOK)
}

// Head 获取分片临时对象的大小
func Head(g *gin.Context) {
	id := g.Param("name")
	bt, ok := pool.Cache.HasGet(id)
	if !ok {
		g.Status(http.StatusNotFound)
		return
	}
	ti, ok := util.GobDecodeGen[entity.TempInfo](bt)
	if !ok {
		g.Status(http.StatusInternalServerError)
		return
	}
	fi, err := os.Stat(filepath.Join(pool.Config.TempPath, id))
	if err != nil {
		if os.IsNotExist(err) {
			g.Status(http.StatusNotFound)
			return
		}
		util.LogErr(err)
		g.Status(http.StatusInternalServerError)
		return
	}
	// fi may have aligned padding if upload has finished
	g.Header("Size", util.ToString(xmath.MinNumber(fi.Size(), ti.Size)))
	g.Status(http.StatusOK)
}

// Get 获取临时对象分片
func Get(g *gin.Context) {
	req := struct {
		Name string `uri:"name" binding:"required"`
		Size int64  `form:"size" binding:"gte=1"`
	}{}
	if err := entity.BindAll(g, &req, binding.Uri, binding.Query); err != nil {
		g.Status(http.StatusBadRequest)
		return
	}
	if err := service.GetTemp(req.Name, req.Size, g.Writer); err != nil {
		util.LogErr(err)
		g.Status(http.StatusNotFound)
		return
	}
	g.Status(http.StatusOK)
}
