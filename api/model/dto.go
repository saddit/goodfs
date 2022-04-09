package model

import (
	"goodfs/api/repository/metadata"

	"github.com/gin-gonic/gin"
)

type PutResp struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

type PutReq struct {
	Name string `uri:"name" binding:"required"`
	Hash string `header:"Digit" binding:"required"`
}

type GetReq struct {
	Name    string `uri:"name" binding:"required"`
	Version int    `form:"version"`
}

func (p *PutReq) Bind(c *gin.Context) error {
	if err := c.ShouldBindUri(p); err != nil {
		return err
	}
	if err := c.ShouldBindHeader(p); err != nil {
		return err
	}
	return nil
}

func (p *GetReq) Bind(c *gin.Context) error {
	if err := c.ShouldBindUri(p); err != nil {
		return err
	}
	p.Version = metadata.VerModeLast
	_ = c.ShouldBindQuery(p)
	return nil
}
