package big

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/service"
	"common/logs"
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

//Post 生成大文件上传的Token
func (bc *BigObjectsController) Post(g *gin.Context) {
	req := g.Value("BigPostReq").(*entity.BigPostReq)
	ips := logic.NewDiscovery().SelectDataServer(pool.Balancer, pool.Config.Rs.AllShards())
	if len(ips) == 0 {
		response.ServiceUnavailableMsg("no available servers", g)
		return
	}
	stream, e := service.NewRSResumablePutStream(ips, req.Name, req.Hash, req.Size)
	if e != nil {
		response.FailErr(e, g)
		return
	}
	defer stream.Close()
	response.CreatedHeader(gin.H{
		"Location": "/big/" + url.PathEscape(stream.Token()),
	}, g)
}

//Head 大文件已上传大小
func (bc *BigObjectsController) Head(g *gin.Context) {
	token, _ := url.PathUnescape(g.Param("token"))
	stream, e := service.NewRSResumablePutStreamFromToken(token)
	if e != nil {
		response.BadRequestErr(e, g)
		return
	}
	defer stream.Close()
	size := stream.CurrentSize()
	if size == -1 {
		response.NotFound(g)
		return
	}
	response.OkHeader(gin.H{
		"Content-Length": util.ToString(size),
	}, g)
}

//Patch 上传大文件
func (bc *BigObjectsController) Patch(g *gin.Context) {
	var req entity.BigPutReq
	if e := req.Bind(g); e != nil {
		response.BadRequestErr(e, g)
		return
	}
	stream, e := service.NewRSResumablePutStreamFromToken(req.Token)
	if e != nil {
		g.JSON(http.StatusBadRequest, gin.H{"msg": e.Error()})
		return
	}
	defer stream.Close()
	curSize := stream.CurrentSize()
	if curSize != req.Range.Value().First {
		response.Exec(g).Status(http.StatusRequestedRangeNotSatisfiable).Abort()
		return
	}
	bufSize := int64(pool.Config.Rs.BlockSize())
	for {
		n, err := io.CopyN(stream, g.Request.Body, bufSize)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			response.FailErr(err, g)
			return
		}
		curSize += n
		//大于预先确定的大小 则属于异常访问
		if curSize > stream.Size {
			_ = stream.Commit(false)
			logs.Std().Infoln("resumable put exceed size")
			response.Exec(g).Status(http.StatusForbidden).Abort()
			return
		}
		//上传未完成 中断
		if n != bufSize && curSize != stream.Size {
			response.Exec(g).Status(http.StatusPartialContent)
			return
		}
		//上传未完成 继续
		if curSize != stream.Size {
			continue
		}
		if pool.Config.Checksum {
			getStream, err := service.NewRSGetStream(stream.Size, stream.Hash, stream.Locates)
			if err != nil {
				response.FailErr(err, g)
				return
			}
			hash := crypto.SHA256IO(getStream)
			if hash != stream.Hash {
				if err = stream.Commit(false); err != nil {
					logs.Std().Error(err)
				}
				response.Exec(g).Status(http.StatusForbidden).Abort()
				return
			}
		}
		if err = stream.Commit(true); err != nil {
			response.FailErr(err, g)
			return
		}
		verNum, err := bc.metaService.SaveMetadata(&entity.Metadata{
			Name: stream.Name,
			Versions: []*entity.Version{{
				Hash:   stream.Hash,
				Size:   stream.Size,
				Locate: stream.Locates,
			}},
		})
		if err != nil {
			response.FailErr(err, g)
			return
		}
		response.OkJson(&entity.PutResp{
			Name:    stream.Name,
			Version: verNum,
		}, g)
	}
}
