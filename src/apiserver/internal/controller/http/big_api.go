package http

import (
	"apiserver/config"
	"apiserver/internal/entity"
	"apiserver/internal/usecase"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/repo"
	"apiserver/internal/usecase/service"
	"common/graceful"
	"common/response"
	"common/util"
	"common/util/crypto"
	"fmt"

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
	r.POST("/big/:name", bc.Post)
	r.HEAD("/big/:token", bc.Head)
	r.PATCH("/big/:token", bc.Patch)
}

// Post prepare a resumable uploading
func (bc *BigObjectsController) Post(g *gin.Context) {
	var req entity.BigPostReq
	if err := req.Bind(g); err != nil {
		response.BadRequestErr(err, g).Abort()
		return
	}
	req.Ext = util.GetFileExtOrDefault(req.Name, false, "bytes")
	bucket, err := bc.bucketRepo.Get(req.Bucket)
	if err != nil {
		response.FailErr(err, g)
		return
	}
	if bucket.Readonly {
		response.BadRequestMsg("bucket is readonly", g)
		return
	}
	// if bucket enforce compress
	if bucket.Compress {
		req.Compress = true
	}
	// configure by bucket config
	conf := bucket.MakeConf(&pool.Config.Object).ReedSolomon
	// generate a unique hash as version hash
	uniqueHash := bc.objectService.UniqueHash(req.Hash, entity.ECReedSolomon, conf.DataShards, conf.ParityShards, req.Compress)
	// filter duplicate
	locates, ok := bc.objectService.LocateObject(uniqueHash, conf.AllShards())
	if ok {
		// finish upload
		object := &entity.Version{
			Hash:          uniqueHash,
			Size:          req.Size,
			Compress:      req.Compress,
			Locate:        locates,
			DataShards:    conf.DataShards,
			ParityShards:  conf.ParityShards,
			ShardSize:     conf.ShardSize(req.Size),
			StoreStrategy: entity.ECReedSolomon,
		}
		verNum, err := bc.finishUpload(req.Name, req.Bucket, object, &conf)
		if err != nil {
			response.FailErr(err, g)
			return
		}
		response.OkJson(&entity.PutResp{
			Name:    req.Name,
			Bucket:  req.Bucket,
			Version: verNum,
		}, g)
		return
	}
	ips := logic.NewDiscovery().SelectDataServer(pool.Balancer, conf.AllShards())
	if len(ips) == 0 {
		response.ServiceUnavailableMsg("no available servers", g)
		return
	}
	stream, e := service.NewRSResumablePutStream(&service.StreamOption{
		Hash:     uniqueHash,
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
		response.Exec(g).
			Fail(http.StatusRequestedRangeNotSatisfiable, fmt.Sprintf("current size is %d, but range is %v", curSize, req.Range))
		return
	}
	// copy from request body
	n, err := io.Copy(stream, g.Request.Body)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		response.FailErr(err, g)
		return
	}
	curSize += n
	if curSize < stream.Size {
		// not read enough, see as interrupted
		response.Exec(g).Status(http.StatusPartialContent).
			Header(gin.H{"Content-Length": curSize})
		return
	}
	// if over the intended size, its invalid
	if curSize > stream.Size {
		util.LogErr(stream.Commit(false))
		response.Exec(g).Fail(http.StatusForbidden, "file exceed the intended size")
		return
	}
	// if curSize equals expected size
	verNum, err := bc.finishUpload(stream.Name, stream.Bucket, &entity.Version{
		Hash:          stream.Hash,
		Size:          stream.Size,
		Locate:        stream.Locates,
		DataShards:    stream.Config.DataShards,
		ParityShards:  stream.Config.ParityShards,
		ShardSize:     stream.Config.ShardSize(stream.Size),
		StoreStrategy: entity.ECReedSolomon,
	}, stream.Config)
	if err != nil {
		util.LogErr(stream.Commit(false))
		response.FailErr(err, g)
		return
	}
	if err = stream.Commit(true); err != nil {
		response.FailErr(err, g)
		return
	}
	response.OkJson(&entity.PutResp{
		Name:    stream.Name,
		Bucket:  stream.Bucket,
		Version: verNum,
	}, g)
}

func (bc *BigObjectsController) finishUpload(metaName, bucketName string, v *entity.Version, conf *config.RsConfig) (verNum int32, err error) {
	// validate digest
	if pool.Config.Object.Checksum {
		getStream := service.NewRSTempStream(&service.StreamOption{
			Hash:    v.Hash,
			Size:    v.Size,
			Locates: v.Locate,
		}, conf)
		hash := crypto.SHA256IO(getStream)
		if hash != v.Hash {
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
		metadata, inner = bc.metaService.GetMetadata(metaName, bucketName, int32(entity.VerModeNot), true)
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
		bucket, inner = bc.bucketRepo.Get(bucketName)
		if inner != nil {
			dg.Error(inner)
		}
	}()
	// wait
	if err = dg.WaitUntilError(); err != nil {
		return
	}
	if metadata != nil {
		// update metadata
		verNum, err = bc.metaService.AddVersion(metadata.Name, metadata.Bucket, v)
		if err != nil {
			return
		}
		if metadata.Total > 0 && !bucket.Versioning || metadata.Total >= bucket.VersionRemains {
			go func() {
				defer graceful.Recover()
				// if not err, delete first version
				inner := bc.metaService.RemoveVersion(metadata.Name, metadata.Bucket, int32(metadata.FirstVersion))
				util.LogErrWithPre("remove first version err", inner)
			}()
		}
	} else {
		// add metadata
		verNum, err = bc.metaService.SaveMetadata(&entity.Metadata{
			Name:     metaName,
			Bucket:   bucketName,
			Versions: []*entity.Version{v},
		})
		if err != nil {
			return
		}
	}
	return
}
