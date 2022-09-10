package objects

import (
	"apiserver/internal/entity"
	"common/response"
	"common/util"
	"github.com/gin-gonic/gin"
)

func ValidatePut(g *gin.Context) {
	var req entity.PutReq
	if err := req.Bind(g); err != nil {
		response.BadRequestErr(err, g).Abort()
		return
	} else if g.Request.ContentLength == 0 {
		response.BadRequestMsg("empty request", g).Abort()
		return
	} else {
		defer g.Set("PutReq", &req)
	}

	req.FileName = req.Hash
	if ext, ok := util.GetFileExt(req.Name, false); ok {
		req.Ext = ext
	} else {
		req.Ext = "bytes"
	}
	//FIXME: 此处直接使用没有验证过的Hash去重文件
	if loc, ok := ObjectService.LocateObject(req.Hash); ok {
		req.Locate = loc
	}
}
