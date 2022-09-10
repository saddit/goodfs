package objects

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase"
	"common/response"
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

	if metaData, ok := MetaService.GetMetadata(req.Name, req.Version); ok {
		metaVer := metaData.Versions[0]
		stream, e := ObjectService.GetObject(metaVer)
		defer stream.Close()
		if e == usecase.ErrBadRequest {
			response.BadRequestErr(e, c).Abort()
		} else if e != nil {
			response.ServiceUnavailableErr(e, c).Abort()
		} else {
			if tp, ok := req.Range.Get(); ok {
				if _, e = stream.Seek(tp.First, io.SeekCurrent); e != nil {
					response.FailErr(e, c)
					return
				}
			}
			_, e = io.CopyBuffer(c.Writer, stream, make([]byte, 2048))
			if e == nil {
				response.Ok(c)
			} else {
				response.FailErr(e, c)
			}
		}
	} else {
		response.NotFound(c).Abort()
	}
}
