package big

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"goodfs/apiserver/global"
	"goodfs/apiserver/model"
	"goodfs/apiserver/model/meta"
	"goodfs/apiserver/service"
	"goodfs/apiserver/service/objectstream"
	"goodfs/lib/util"
	"io"
	"net/http"
	"net/url"
)

//Post 生成大文件上传的Token
func Post(g *gin.Context) {
	req := g.Value("BigPostReq").(*model.BigPostReq)
	ips := service.SelectDataServer(global.Config.Rs.AllShards())
	if len(ips) == 0 {
		g.AbortWithStatus(http.StatusServiceUnavailable)
	}
	stream, e := objectstream.NewRSResumablePutStream(ips, req.Name, req.Hash, req.Size)
	if e != nil {
		util.AbortInternalError(g, e)
		return
	}
	defer stream.Close()
	g.Header("Location", fmt.Sprintf("/big/%s", url.PathEscape(stream.Token())))
	g.Status(http.StatusCreated)
}

//Head 大文件已上传大小
func Head(g *gin.Context) {
	token, e := url.PathUnescape(g.Param("token"))
	if e != nil {
		g.Status(http.StatusBadRequest)
		return
	}
	stream, e := objectstream.NewRSResumablePutStreamFromToken(token)
	if e != nil {
		g.Status(http.StatusBadRequest)
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
	var req model.BigPutReq
	if e := req.Bind(g); e != nil {
		g.AbortWithStatus(http.StatusBadRequest)
		return
	}
	stream, e := objectstream.NewRSResumablePutStreamFromToken(req.Token)
	if e != nil {
		util.AbortInternalError(g, e)
		return
	}
	curSize := stream.CurrentSize()
	if curSize != req.Range.Value().First {
		g.AbortWithStatus(http.StatusRequestedRangeNotSatisfiable)
		return
	}
	bufSize := int64(global.Config.Rs.BlockSize())
	for {
		n, e := io.CopyN(stream, g.Request.Body, bufSize)
		if e != nil && e != io.EOF && e != io.ErrUnexpectedEOF {
			_ = stream.Close()
			util.AbortInternalError(g, e)
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
		//上传完成
		if curSize == stream.Size {
			if global.Config.EnableHashCheck && !checkSum(stream.Size, stream.Hash, stream.Locates) {
				if e = stream.Commit(false); e != nil {
					util.AbortInternalError(g, e)
				} else {
					g.AbortWithStatus(http.StatusForbidden)
				}
				return
			}
			if e = stream.Commit(true); e != nil {
				util.AbortInternalError(g, e)
				return
			} else {
				var verNum int32
				verNum, e = service.SaveMetadata(&meta.Data{
					Name: stream.Name,
					Versions: []*meta.Version{{
						Hash:   stream.Hash,
						Size:   stream.Size,
						Locate: stream.Locates,
					}},
				})
				if e != nil {
					util.AbortInternalError(g, e)
					return
				} else {
					g.JSON(http.StatusOK, model.PutResp{
						Name:    stream.Name,
						Version: verNum,
					})
					return
				}
			}
		} else if n != bufSize {
			//上传未完成 中断
			stream.Close()
			g.Status(http.StatusPartialContent)
			return
		}
	}
}

func checkSum(size int64, hash string, locates []string) bool {
	getStream, e := objectstream.NewRSGetStream(size, hash, locates)
	if e != nil {
		log.Errorln(e)
		return false
	}
	result := util.SHA256Hash(getStream)
	if hash != result {
		return false
	}
	return true
}
