package objects

import (
	"bytes"
	"common/graceful"
	"common/request"
	"common/response"
	"io"
	"net/http"
	"objectserver/internal/entity"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func Put(c *gin.Context) {
	req := &struct {
		Name     string `uri:"name"`
		Compress bool   `form:"compress"`
	}{}
	if err := entity.BindAll(c, req, binding.Uri, binding.Query); err != nil {
		response.FailErr(err, c)
		return
	}
	if service.Exist(req.Name) {
		response.Ok(c)
		return
	}
	var reader io.Reader = c.Request.Body
	var cache bytes.Buffer
	if uint64(c.Request.ContentLength) <= pool.Config.Cache.MaxItemSize.Byte() {
		cache.Grow(int(c.Request.ContentLength))
		reader = io.TeeReader(c.Request.Body, &cache)
	}
	if err := service.Put(req.Name, reader, req.Compress); err != nil {
		response.FailErr(err, c)
		return
	}
	if cache.Len() > 0 {
		go func() {
			defer graceful.Recover()
			pool.Cache.Set(req.Name, cache.Bytes())
		}()
	}
	response.Ok(c)
}

func Delete(c *gin.Context) {
	name := c.Param("name")
	pool.Cache.Delete(name)
	if err := service.Delete(name); err != nil {
		response.FailErr(err, c)
		return
	}
	c.Status(http.StatusNoContent)
}

func Get(c *gin.Context) {
	req := &struct {
		Name     string `uri:"name"`
		Range    string `header:"range"`
		Size     int64  `header:"size" binding:"required"`
		Compress bool   `form:"compress"`
	}{}
	if err := entity.BindAll(c, req, binding.Uri, binding.Query, binding.Header); err != nil {
		response.FailErr(err, c)
		return
	}
	var rg request.Range
	var offset int64
	if ok := rg.ConvertFrom(req.Range); ok {
		offset = rg.FirstBytes().First
	}
	var writer io.Writer = c.Writer
	var buf bytes.Buffer
	if uint64(req.Size) <= pool.Config.Cache.MaxItemSize.Byte() {
		buf.Grow(int(req.Size))
		writer = io.MultiWriter(c.Writer, &buf)
	}
	if err := service.Get(req.Name, offset, req.Size, req.Compress, writer); err != nil {
		response.FailErr(err, c)
		return
	}
	if buf.Len() > 0 {
		go func() {
			defer graceful.Recover()
			pool.Cache.Set(req.Name, buf.Bytes())
		}()
	}
	c.Status(http.StatusOK)
}

func Head(c *gin.Context) {
	fileName := c.Param("name")
	if ok := service.Exist(fileName); ok {
		c.Status(http.StatusOK)
		return
	}
	c.Status(http.StatusNotFound)
}
