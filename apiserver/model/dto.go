package model

import (
	"goodfs/apiserver/repository/metadata"
	"goodfs/lib/util"
	"io"

	"github.com/gin-gonic/gin/binding"

	"github.com/gin-gonic/gin"
)

type SyncTyping string

const (
	SyncInsert SyncTyping = "insert"
	SyncDelete SyncTyping = "delete"
)

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
}

func (p *PutReq) Bind(c *gin.Context) error {
	if err := util.BindAll(c, p, binding.Header, binding.Uri); err != nil {
		return err
	}
	return nil
}

func (p *GetReq) Bind(c *gin.Context) error {
	p.Version = int32(metadata.VerModeLast)
	if err := util.BindAll(c, p, binding.Query, binding.Uri); err != nil {
		return err
	}
	return nil
}
