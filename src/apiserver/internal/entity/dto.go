package entity

import (
	"common/request"
	"common/response"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io"
	"net/url"
)

type SyncTyping string

type PutResp struct {
	Name    string `json:"name"`
	Version int32  `json:"version"`
}

type PutReq struct {
	Name     string `uri:"name" binding:"required"`
	Hash     string `header:"digest" binding:"required"`
	Ext      string
	Locate   []string
	FileName string
	Body     io.Reader
}

type GetReq struct {
	Name    string `uri:"name" binding:"required"`
	Version int32  `form:"version" binding:"min=0"`
	Range   request.Range
}

type BigPostReq struct {
	Name string `uri:"name" binding:"required"`
	Hash string `header:"digest" binding:"required"`
	Size int64  `header:"size" binding:"required"`
	Ext  string
}

type BigPutReq struct {
	Token string `uri:"token" binding:"required"`
	Range request.Range
}

func (b *BigPostReq) Bind(c *gin.Context) error {
	if err := BindAll(c, b, binding.Header, binding.Uri); err != nil {
		return err
	}
	return nil
}

func (bigPut *BigPutReq) Bind(c *gin.Context) error {
	var err error
	if err = BindAll(c, bigPut, binding.Header, binding.Uri); err != nil {
		return err
	}
	if bigPut.Token, err = url.PathUnescape(bigPut.Token); err != nil {
		return err
	}
	if rangeStr := c.GetHeader("Range"); len(rangeStr) > 0 {
		if ok := bigPut.Range.ConvertFrom(rangeStr); ok {
			return nil
		}
	}
	return fmt.Errorf("require header 'Range'")
}

func (p *PutReq) Bind(c *gin.Context) error {
	if err := BindAll(c, p, binding.Header, binding.Uri); err != nil {
		return err
	}
	return nil
}

func (g *GetReq) Bind(c *gin.Context) error {
	g.Version = int32(VerModeLast)
	if err := BindAll(c, g, binding.Query, binding.Uri); err != nil {
		return err
	}
	if rangeStr := c.GetHeader("Range"); len(rangeStr) > 0 {
		if ok := g.Range.ConvertFrom(rangeStr); !ok {
			return response.NewError(400, "header 'Range' format error")
		}
	}
	return nil
}
