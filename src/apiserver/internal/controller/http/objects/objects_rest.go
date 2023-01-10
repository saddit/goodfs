package objects

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase"
	"common/response"
	"common/util"
	"github.com/gin-gonic/gin"
	"io"
)

type ObjectsController struct {
	objectService usecase.IObjectService
	metaService   usecase.IMetaService
}

func NewObjectsController(obj usecase.IObjectService, meta usecase.IMetaService) *ObjectsController {
	return &ObjectsController{obj, meta}
}

func (oc *ObjectsController) Register(r gin.IRoutes) {
	r.PUT("/objects/:name", ValidatePut(oc.objectService), oc.Put)
	r.GET("/objects/:name", oc.Get)
}

func (oc *ObjectsController) Put(c *gin.Context) {
	req := c.Value("PutReq").(*entity.PutReq)
	req.Body = c.Request.Body
	if c.Request.ContentLength <= 0 {
		response.BadRequestMsg("content-length invalid", c)
		return
	}
	verNum, err := oc.objectService.StoreObject(req, &entity.Metadata{
		Name: req.Name,
		Versions: []*entity.Version{{
			Size:          c.Request.ContentLength,
			Hash:          req.Hash,
			StoreStrategy: req.Store,
		}},
	})

	if err != nil {
		response.FailErr(err, c)
		return
	}

	response.CreatedJson(&entity.PutResp{
		Name:    req.Name,
		Version: verNum,
	}, c)
}

func (oc *ObjectsController) Get(c *gin.Context) {
	var req entity.GetReq
	if e := req.Bind(c); e != nil {
		response.BadRequestErr(e, c)
		return
	}
	// get metadata
	metaData, err := oc.metaService.GetMetadata(req.Name, req.Version)
	if err != nil {
		response.FailErr(err, c).Abort()
		return
	}
	// get object stream
	stream, err := oc.objectService.GetObject(metaData, metaData.Versions[0])
	defer util.CloseAndLog(stream)
	if err != nil {
		response.FailErr(err, c).Abort()
		return
	}
	// try seek
	if tp, ok := req.Range.GetFirstBytes(); ok {
		if _, err = stream.Seek(tp.First, io.SeekCurrent); err != nil {
			response.FailErr(err, c)
			return
		}
	}
	// copy to response
	_, err = io.Copy(c.Writer, stream)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.Ok(c)
}
