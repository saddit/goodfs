package objects

import (
	"bytes"
	"common/request"
	"common/response"
	"common/util"
	"io"
	"net/http"
	"objectserver/internal/entity"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/service"

	"github.com/gin-gonic/gin"
)

func Put(c *gin.Context) {
	fileName := c.Param("name")
	var reader io.Reader = c.Request.Body
	var cache []byte
	if uint64(c.Request.ContentLength) <= pool.Config.Cache.MaxItemSize.Byte() {
		cache = make([]byte, 0, c.Request.ContentLength)
		reader = &entity.BufferTeeReader{Reader: c.Request.Body, Body: bytes.NewBuffer(cache)}
	}
	if err := service.Put(fileName, reader); err != nil {
		response.FailErr(err, c)
		return
	}
	if len(cache) > 0 {
		pool.Cache.Set(fileName, cache)
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
	size := util.ToInt64(c.GetHeader("Size"))
	fileName := c.Param("name")
	var rg request.Range
	var offset int64
	if ok := rg.ConvertFrom(c.GetHeader("Range")); ok {
		offset = rg.FirstBytes().First
	}
	var writer io.Writer = c.Writer
	var buf []byte
	if uint64(size) <= pool.Config.Cache.MaxItemSize.Byte() {
		buf = make([]byte, 0, size)
		writer = &entity.BufferTeeWriter{Writer: c.Writer, Body: bytes.NewBuffer(buf)}
	}
	if err := service.Get(fileName, offset, size, writer); err != nil {
		response.FailErr(err, c)
		return
	}
	if len(buf) > 0 {
		pool.Cache.Set(fileName, buf)
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
