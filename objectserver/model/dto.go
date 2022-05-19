package model

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"goodfs/lib/util"
	"strconv"
	"strings"
)

const (
	SyncInsert = "insert"
)

const TempKeyPrefix = "TempInfo_"

type TempPostReq struct {
	Name string `uri:"name" binding:"required"`
	Size int64  `header:"size" binding:"required"`
}

func (tp *TempPostReq) Bind(c *gin.Context) error {
	if e := util.BindAll(c, tp, binding.Uri, binding.Header); e != nil {
		return e
	}
	return nil
}

type TempInfo struct {
	Name string
	Id   string
	Size int64
}

func (t *TempInfo) ShardIndex() int {
	s := strings.Split(t.Name, ".")
	id, _ := strconv.Atoi(s[1])
	return id
}
