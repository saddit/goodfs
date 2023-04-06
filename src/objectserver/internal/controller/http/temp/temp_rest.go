package temp

import (
	"common/cst"
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
)

func Patch(g *gin.Context) {
	id := g.Param("name")
	ti, ok := service.GetTempInfo(id)
	if !ok {
		response.BadRequestMsg("file has been removed", g)
		return
	}
	// only allow last chuck may not be power of 4KB
	// for reading from network-io, using too big buffer is not wise.
	bufSize := xmath.MinInt(int(g.Request.ContentLength), 2*cst.OS.PageSize)
	if _, err := service.AppendFileAligned(ti.FullPath, g.Request.Body, bufSize); err != nil {
		response.FailErr(err, g)
		return
	}
	response.Ok(g)
}

func Delete(g *gin.Context) {
	id := g.Param("name")
	defer service.RemoveTempInfo(id)
	response.Ok(g)
}

func Post(g *gin.Context) {
	var req entity.TempPostReq
	if err := entity.BindAll(g, &req, binding.Header, binding.Uri); err != nil {
		response.FailErr(err, g)
		return
	}
	tmpInfo := &entity.TempInfo{
		Name:       req.Name,
		Size:       req.Size,
		Id:         service.GenerateTempID(),
		MountPoint: pool.DriverManager.SelectMountPointFallback(pool.Config.BaseMountPoint),
	}
	tmpInfo.FullPath = filepath.Join(tmpInfo.MountPoint, pool.Config.TempPath, tmpInfo.Id)
	if !service.SetTempInfo(tmpInfo) {
		g.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	g.Status(http.StatusOK)
	_, _ = g.Writer.Write(util.StrToBytes(tmpInfo.Id))
}

func Put(g *gin.Context) {
	req := &struct {
		ID       string `uri:"name"`
		Compress bool   `form:"compress"`
	}{}
	if err := entity.BindAll(g, req, binding.Uri, binding.Query); err != nil {
		response.FailErr(err, g)
		return
	}
	ti, ok := service.GetTempInfo(req.ID)
	if !ok {
		response.BadRequestMsg("file has been removed", g)
		return
	}
	if err := service.CommitFile(ti.MountPoint, req.ID, ti.Name, req.Compress); err != nil {
		response.FailErr(err, g)
		return
	}
	service.RemoveTempInfo(req.ID)
	response.Ok(g)
}

// Head 获取分片临时对象的大小
func Head(g *gin.Context) {
	id := g.Param("name")
	ti, ok := service.GetTempInfo(id)
	if !ok {
		g.Status(http.StatusNotFound)
		return
	}
	fi, err := os.Stat(ti.FullPath)
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
	ti, ok := service.GetTempInfo(req.Name)
	if !ok {
		response.BadRequestMsg("file has been removed", g)
		return
	}
	if err := service.GetFile(ti.FullPath, 0, req.Size, g.Writer); err != nil {
		response.FailErr(err, g)
		return
	}
	response.Ok(g)
}
