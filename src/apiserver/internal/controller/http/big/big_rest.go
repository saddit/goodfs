package big

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/repo"
	"apiserver/internal/usecase/service"
	"common/graceful"
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
	bucketRepo    repo.IBucketRepo
}

func NewBigObjectsController(obj usecase.IObjectService, meta usecase.IMetaService, buk repo.IBucketRepo) *BigObjectsController {
	return &BigObjectsController{obj, meta, buk}
}

func (bc *BigObjectsController) Register(r gin.IRoutes) {
	r.POST("/big/:name", FilterDuplicates(bc.objectService), bc.Post)
	r.HEAD("/big/:token", bc.Head)
	r.PATCH("/big/:token", bc.Patch)
}

// Post prepare a resumable uploading
func (bc *BigObjectsController) Post(g *gin.Context) {
	req := g.Value("BigPostReq").(*entity.BigPostReq)
	bucket, err := bc.bucketRepo.Get(req.Bucket)
	if err != nil {
		response.FailErr(err, g)
		return
	}
	if bucket.Readonly {
		response.BadRequestMsg("bucket is readonly", g)
		return
	}
	conf := pool.Config.Rs
	// if bucket enforce compress
	if bucket.Compress {
		req.Compress = true
	}
	// if bucket enforce store strategy
	if bucket.StoreStrategy == entity.ECReedSolomon {
		conf.DataShards = bucket.DataShards
		conf.ParityShards = bucket.ParityShards
	}
	ips := logic.NewDiscovery().SelectDataServer(pool.Balancer, conf.AllShards())
	if len(ips) == 0 {
		response.ServiceUnavailableMsg("no available servers", g)
		return
	}
	stream, e := service.NewRSResumablePutStream(&service.StreamOption{
		Hash:     req.Hash,
		Name:     req.Name,
		Size:     req.Size,
		Bucket:   req.Bucket,
		Compress: req.Compress,
		Locates:  ips,
	}, &conf)
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
	for curSize < stream.Size {
		n, err := io.CopyN(stream, g.Request.Body, bufSize)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			response.FailErr(err, g)
			return
		}
		curSize += n
		if curSize < stream.Size {
			if n != bufSize {
				// not read enough, see as interrupted
				response.Exec(g).Status(http.StatusPartialContent)
				return
			}
		}
	}
	// over the intended size, its invalid
	if curSize > stream.Size {
		util.LogErr(stream.Commit(false))
		response.Exec(g).Fail(http.StatusForbidden, "file exceed the intended size")
		return
	}
	verNum, err := bc.finishUpload(stream)
	if err != nil {
		response.FailErr(err, g)
		return
	}
	response.OkJson(&entity.PutResp{
		Name:    stream.Name,
		Bucket:  stream.Bucket,
		Version: verNum,
	}, g)
}

func (bc *BigObjectsController) finishUpload(stream *service.RSResumablePutStream) (verNum int32, err error) {
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
			err = response.NewError(http.StatusForbidden, "signature authentication failure")
			return
		}
	}
	dg := util.NewDoneGroup()
	defer dg.Close()
	// get metadata if exist
	var metadata *entity.Metadata
	dg.Todo()
	go func() {
		defer dg.Done()
		var inner error
		metadata, inner = bc.metaService.GetMetadata(stream.Name, stream.Bucket, int32(entity.VerModeNot), true)
		if err != nil && !response.CheckErrStatus(404, inner) {
			dg.Error(inner)
		}
	}()
	// get bucket
	var bucket *entity.Bucket
	dg.Todo()
	go func() {
		defer dg.Done()
		var inner error
		bucket, inner = bc.bucketRepo.Get(stream.Bucket)
		if inner != nil {
			dg.Error(inner)
		}
	}()
	// wait
	if err = dg.WaitUntilError(); err != nil {
		return
	}
	// update metadata
	verData := &entity.Version{
		Hash:          stream.Hash,
		Size:          stream.Size,
		Locate:        stream.Locates,
		DataShards:    stream.Config.DataShards,
		ParityShards:  stream.Config.ParityShards,
		ShardSize:     stream.Config.ShardSize(stream.Size),
		StoreStrategy: entity.ECReedSolomon,
	}
	if metadata != nil {
		verNum, err = bc.metaService.AddVersion(stream.Name, stream.Bucket, verData)
		if err != nil {
			return
		}
		if metadata.Total > 0 && !bucket.Versioning || metadata.Total >= bucket.VersionRemains {
			go func() {
				defer graceful.Recover()
				// if not err, delete first version
				inner := bc.metaService.RemoveVersion(stream.Name, stream.Bucket, int32(metadata.FirstVersion))
				util.LogErrWithPre("remove first version err", inner)
			}()
		}
	} else {
		// update metadata
		verNum, err = bc.metaService.SaveMetadata(&entity.Metadata{
			Name:     stream.Name,
			Bucket:   stream.Bucket,
			Versions: []*entity.Version{verData},
		})
		if err != nil {
			return
		}
	}
	// commit upload
	err = stream.Commit(true)
	return
}
