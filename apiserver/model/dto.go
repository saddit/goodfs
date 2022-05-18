package model

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"goodfs/apiserver/repository/metadata"
	"goodfs/lib/util"
	"io"
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
	Version int32  `form:"version"`
	Range   Range
}

type BigPostReq struct {
	Name string `uri:"name" binding:"required"`
	Hash string `header:"digest" binding:"required"`
	Size int64  `header:"size" binding:"required"`
}

type BigPutReq struct {
	Token string `uri:"token" binding:"required"`
	Range Range
}

func (b *BigPostReq) Bind(c *gin.Context) error {
	if err := util.BindAll(c, b, binding.Header, binding.Uri); err != nil {
		return err
	}
	return nil
}

func (bigPut *BigPutReq) Bind(c *gin.Context) error {
	if err := util.BindAll(c, bigPut, binding.Header, binding.Uri); err != nil {
		return err
	}
	if rangeStr := c.GetHeader("Range"); len(rangeStr) > 0 {
		var r Range
		if ok := r.convertFrom(rangeStr); ok {
			bigPut.Range = r
			return nil
		}
	}
	return fmt.Errorf("require header 'Range'")
}

func (p *PutReq) Bind(c *gin.Context) error {
	if err := util.BindAll(c, p, binding.Header, binding.Uri); err != nil {
		return err
	}
	return nil
}

func (g *GetReq) Bind(c *gin.Context) error {
	g.Version = int32(metadata.VerModeLast)
	if err := util.BindAll(c, g, binding.Query, binding.Uri); err != nil {
		return err
	}
	if rangeStr := c.GetHeader("Range"); len(rangeStr) > 0 {
		var r Range
		if ok := r.convertFrom(rangeStr); ok {
			g.Range = r
		}
	}
	return nil
}
