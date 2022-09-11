package objects

import (
	"apiserver/internal/entity"
	"common/response"
	"common/util"
	"github.com/gin-gonic/gin"
	"io"
)

func Put(c *gin.Context) {
	req := c.Value("PutReq").(*entity.PutReq)
	req.Body = c.Request.Body
	verNum, err := ObjectService.StoreObject(req, &entity.Metadata{
		Name: req.Name,
		Versions: []*entity.Version{{
			Size: c.Request.ContentLength,
			Hash: req.Hash,
		}},
	})

	if err != nil {
		response.FailErr(err, c)
		return
	}

	response.OkJson(&entity.PutResp{
		Name:    req.Name,
		Version: verNum,
	}, c)
}

func Get(c *gin.Context) {
	var req entity.GetReq
	if e := req.Bind(c); e != nil {
		response.BadRequestErr(e, c)
		return
	}
	// get metadata
	metaData, err := MetaService.GetMetadata(req.Name, req.Version)
	if err != nil {
		response.FailErr(err, c).Abort()
		return
	}
	// get object stream
	stream, err := ObjectService.GetObject(metaData, metaData.Versions[0])
	defer util.CloseAndLog(stream)
	if err != nil {
		response.FailErr(err, c).Abort()
		return
	}
	// try seek
	if tp, ok := req.Range.Get(); ok {
		if _, err = stream.Seek(tp.First, io.SeekCurrent); err != nil {
			response.FailErr(err, c)
			return
		}
	}
	// copy to response
	_, err = io.CopyBuffer(c.Writer, stream, make([]byte, 2048))
	if err == nil {
		response.FailErr(err, c)
		return
	}
	response.Ok(c)
}
