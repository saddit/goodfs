package big

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/service"
	"common/response"
	"common/util"
	"common/util/crypto"

	"apiserver/internal/usecase/logic"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

type BigObjectsController struct {
	objectService usecase.IObjectService
	metaService   usecase.IMetaService
}

func NewBigObjectsController(obj usecase.IObjectService, meta usecase.IMetaService) *BigObjectsController {
	return &BigObjectsController{obj, meta}
}

func (bc *BigObjectsController) Register(r gin.IRoutes) {
	r.POST("/big/:name", FilterDuplicates(bc.objectService), bc.Post)
	r.HEAD("/big/:token", bc.Head)
	r.PATCH("/big/:token", bc.Patch)
}

// Post prepare a resumable uploading
func (bc *BigObjectsController) Post(g *gin.Context) {
	req := g.Value("BigPostReq").(*entity.BigPostReq)
	ips := logic.NewDiscovery().SelectDataServer(pool.Balancer, pool.Config.Rs.AllShards())
	if len(ips) == 0 {
		response.ServiceUnavailableMsg("no available servers", g)
		return
	}
	stream, e := service.NewRSResumablePutStream(&service.StreamOption{
		Hash:    req.Hash,
		Name:    req.Name,
		Size:    req.Size,
		Locates: ips,
	}, &pool.Config.Rs)
	if e != nil {
		response.FailErr(e, g)
		return
	}
	defer stream.Close()
	response.CreatedHeader(gin.H{
		"Accept-Ranges": "bytes",
		"Min-Part-Size": stream.Config.BlockSize(),
		"Location":      "/v1/big/" + url.PathEscape(stream.Token()),
	}, g)
}

// Head uploaded information
func (bc *BigObjectsController) Head(g *gin.Context) {
	token, _ := url.PathUnescape(g.Param("token"))
	stream, e := service.NewRSResumablePutStreamFromToken(token)
	if e != nil {
		response.BadRequestErr(e, g)
		return
	}
	defer stream.Close()
	size, err := stream.CurrentSize()
	if err != nil {
		response.FailErr(err, g)
		return
	}
	response.OkHeader(gin.H{
		"Accept-Ranges":  "bytes",
		"Min-Part-Size":  stream.Config.BlockSize(),
		"Content-Length": size,
	}, g)
}

// Patch upload partial fo file
func (bc *BigObjectsController) Patch(g *gin.Context) {
	var req entity.BigPutReq
	if err := req.Bind(g); err != nil {
		response.BadRequestErr(err, g)
		return
	}
	stream, err := service.NewRSResumablePutStreamFromToken(req.Token)
	if err != nil {
		response.BadRequestErr(err, g)
		return
	}
	defer stream.Close()
	curSize, err := stream.CurrentSize()
	if err != nil {
		response.FailErr(err, g)
		return
	}
	if curSize != req.Range.FirstBytes().First {
		g.Status(http.StatusRequestedRangeNotSatisfiable)
		return
	}
	bufSize := int64(stream.Config.BlockSize())
	for {
		n, err := io.CopyN(stream, g.Request.Body, bufSize)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			response.FailErr(err, g)
			return
		}
		curSize += n
		if curSize > stream.Size {
			// over the intended size, its invalid
			util.LogErr(stream.Commit(false))
			response.Exec(g).Fail(http.StatusForbidden, "file exceed the intended size")
			return
		} else if curSize < stream.Size {
			if n != bufSize {
				// not read enough, see as interrupted
				response.Exec(g).Status(http.StatusPartialContent)
				return
			}
			// not finish yet, continue read from request body
			continue
		}
		// validate digest
		if pool.Config.Checksum {
			getStream := service.NewRSTempStream(&service.StreamOption{
				Hash:    stream.Hash,
				Size:    stream.Size,
				Locates: stream.Locates,
			}, stream.Config)
			hash := crypto.SHA256IO(getStream)
			if hash != stream.Hash {
				util.LogErr(stream.Commit(false))
				response.Exec(g).Fail(http.StatusForbidden, "signature authentication failure")
				return
			}
		}
		// update metadata
		verNum, err := bc.metaService.SaveMetadata(&entity.Metadata{
			Name: stream.Name,
			Versions: []*entity.Version{{
				Hash:          stream.Hash,
				Size:          stream.Size,
				Locate:        stream.Locates,
				DataShards:    stream.Config.DataShards,
				ParityShards:  stream.Config.ParityShards,
				ShardSize:     stream.Config.ShardSize(stream.Size),
				StoreStrategy: entity.ECReedSolomon,
			}},
		})
		if err != nil {
			response.FailErr(err, g)
			return
		}
		// commit upload
		if err = stream.Commit(true); err != nil {
			response.FailErr(err, g)
			return
		}
		response.OkJson(&entity.PutResp{
			Name:    stream.Name,
			Version: verNum,
		}, g)
	}
}
