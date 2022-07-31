package big

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/service"
	"common/util"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"io"
	"net/http"
	"net/url"
)

//Post 生成大文件上传的Token
func Post(g *gin.Context) {
	req := g.Value("BigPostReq").(*entity.BigPostReq)
	ips := service.SelectDataServer(pool.Balancer, pool.Config.Rs.AllShards())
	if len(ips) == 0 {
		g.AbortWithStatus(http.StatusServiceUnavailable)
	}
	stream, e := service.NewRSResumablePutStream(ips, req.Name, req.Hash, req.Size)
	if e != nil {
		AbortInternalError(g, e)
		return
	}
	defer stream.Close()
	g.Header("Location", "/big/"+url.PathEscape(stream.Token()))
	g.Status(http.StatusCreated)
}

//Head 大文件已上传大小
func Head(g *gin.Context) {
	token, _ := url.PathUnescape(g.Param("token"))
	stream, e := service.NewRSResumablePutStreamFromToken(token)
	if e != nil {
		g.JSON(http.StatusBadRequest, gin.H{"msg": e.Error()})
		return
	}
	defer stream.Close()
	size := stream.CurrentSize()
	if size == -1 {
		g.Status(http.StatusNotFound)
	} else {
		g.Header("Content-Length", util.ToString(size))
	}
}

//Patch 上传大文件
func Patch(g *gin.Context) {
	var req entity.BigPutReq
	if e := req.Bind(g); e != nil {
		g.AbortWithStatus(http.StatusBadRequest)
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
		g.AbortWithStatus(http.StatusRequestedRangeNotSatisfiable)
		return
	}
	bufSize := int64(pool.Config.Rs.BlockSize())
	for {
		n, e := io.CopyN(stream, g.Request.Body, bufSize)
		if e != nil && e != io.EOF && e != io.ErrUnexpectedEOF {
			AbortInternalError(g, e)
			return
		}
		curSize += n
		//大于预先确定的大小 则属于异常访问
		if curSize > stream.Size {
			_ = stream.Commit(false)
			log.Infoln("resumable put exceed size")
			g.AbortWithStatus(http.StatusForbidden)
			return
		}
		//上传未完成 中断
		if n != bufSize && curSize != stream.Size {
			g.Status(http.StatusPartialContent)
			return
		}
		//上传完成
		if curSize == stream.Size {
			if pool.Config.EnableHashCheck {
				getStream, e := service.NewRSGetStream(stream.Size, stream.Hash, stream.Locates)
				if e != nil {
					AbortInternalError(g, e)
					return
				}
				hash := util.SHA256Hash(getStream)
				if hash != stream.Hash {
					if e = stream.Commit(false); e != nil {
						log.Println(e)
					}
					g.AbortWithStatus(http.StatusForbidden)
					return
				}
			}
			if e = stream.Commit(true); e != nil {
				AbortInternalError(g, e)
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
					AbortInternalError(g, e)
				} else {
					g.JSON(http.StatusOK, entity.PutResp{
						Name:    stream.Name,
						Version: verNum,
					})
				}
			}
			return
		}
	}
}
