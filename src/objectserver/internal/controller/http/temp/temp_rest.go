package temp

import (
	"common/cache"
	"common/response"
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
	// only allow last chuck may not be power of 4KB
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
		response.FailErr(err, g)
		return
	}
	tmpInfo := &entity.TempInfo{
		Name: req.Name,
		Size: req.Size,
		Id:   entity.TempKeyPrefix + uuid.NewString(),
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
			response.FailErr(err, g)
			return
		}
	} else {
		response.BadRequestMsg("file has been removed", g)
		return
	}
	pool.Cache.Delete(id)
	response.Ok(g)
}

// Head 获取分片临时对象的大小
func Head(g *gin.Context) {
	id := g.Param("name")
	bt, ok := pool.Cache.HasGet(id)
	if !ok {
		g.Status(http.StatusNotFound)
		return
	}
	var ti entity.TempInfo
	if ok = util.GobDecode(bt, &ti); !ok {
		g.Status(http.StatusInternalServerError)
		return
	}
	fi, err := os.Stat(filepath.Join(pool.Config.TempPath, id))
	if os.IsNotExist(err) {
		response.OkHeader(gin.H{"Size": 0}, g)
		return
	}
	if err != nil {
		response.FailErr(err, g)
		return
	}
	// fi may have aligned padding if upload has finished
	response.OkHeader(gin.H{
		"Size": xmath.MinNumber(fi.Size(), ti.Size),
	}, g)
}

// Get 获取临时对象分片
func Get(g *gin.Context) {
	req := struct {
		Name string `uri:"name" binding:"required"`
		Size int64  `header:"size" binding:"gte=1"`
	}{}
	if err := entity.BindAll(g, &req, binding.Uri, binding.Header); err != nil {
		response.BadRequestErr(err, g)
		return
	}
	if err := service.GetTemp(req.Name, req.Size, g.Writer); err != nil {
		response.FailErr(err, g)
		return
	}
	response.Ok(g)
}
