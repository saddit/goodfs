package big

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/service"
	"common/logs"
	"common/response"
	"common/util"

	"apiserver/internal/usecase/logic"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/url"
)

//Post 生成大文件上传的Token
func Post(g *gin.Context) {
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
func Head(g *gin.Context) {
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
	} else {
		response.OkHeader(gin.H{
			"Content-Length": util.ToString(size),
		}, g)
	}
}

//Patch 上传大文件
func Patch(g *gin.Context) {
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
		n, e := io.CopyN(stream, g.Request.Body, bufSize)
		if e != nil && e != io.EOF && e != io.ErrUnexpectedEOF {
			response.FailErr(e, g)
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
		//上传完成
		if curSize == stream.Size {
			if pool.Config.EnableHashCheck {
				getStream, e := service.NewRSGetStream(stream.Size, stream.Hash, stream.Locates)
				if e != nil {
					response.FailErr(e, g)
					return
				}
				hash := util.SHA256Hash(getStream)
				if hash != stream.Hash {
					if e = stream.Commit(false); e != nil {
						logs.Std().Error(e)
					}
					response.Exec(g).Status(http.StatusForbidden).Abort()
					return
				}
			}
			if e = stream.Commit(true); e != nil {
				response.FailErr(e, g)
			} else {
				var verNum int32
				verNum, e = MetaService.SaveMetadata(&entity.Metadata{
					Name: stream.Name,
					Versions: []*entity.Version{{
						Hash:   stream.Hash,
						Size:   stream.Size,
						Locate: stream.Locates,
					}},
				})
				if e != nil {
					response.FailErr(e, g)
				} else {
					response.OkJson(&entity.PutResp{
						Name:    stream.Name,
						Version: verNum,
					}, g)
				}
			}
			return
		}
	}
}
