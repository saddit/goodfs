package service

import (
	"apiserver/internal/entity"
	. "apiserver/internal/usecase"
	"apiserver/internal/usecase/pool"
	"bufio"
	"common/util"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type ObjectService struct {
	metaService IMetaService
}

func NewObjectService(s IMetaService) *ObjectService {
	return &ObjectService{s}
}

// LocateObject 根据Hash定位所有分片位置
func (o *ObjectService) LocateObject(hash string) ([]string, bool) {
	//初始化一个消息接收方（无交换机直接入队）
	conm, e := pool.Amqp.NewConsumer()
	if e != nil {
		return nil, false
	}
	defer conm.Close()
	conm.DeleteUnused = true

	if ch, ok := conm.Consume(); ok {
		//发送定位请求
		for i := 0; i < pool.Config.Rs.AllShards(); i++ {
			pool.AmqpTemplate.PublishDirect("dataServers", "", amqp.Publishing{
				ReplyTo: conm.QueName,
				Body:    []byte(fmt.Sprintf("%s.%d", hash, i)),
			})
		}
		locates := make([]string, pool.Config.Rs.AllShards())
		cnt := 0
		for cnt < pool.Config.Rs.AllShards() {
			select {
			case <-time.After(1 * time.Second):
				log.Warnln("Locate message timeout")
				return locates, cnt == pool.Config.Rs.AllShards()
			case resp := <-ch:
				cnt++
				shardName := resp.Type
				idx, _ := strconv.Atoi(strings.Split(shardName, ".")[1])
				ip := string(resp.Body)
				locates[idx] = ip
			}
		}
		return locates, cnt == pool.Config.Rs.AllShards()
	}
	return nil, false
}

func (o *ObjectService) StoreObject(req *entity.PutReq, md *entity.MetaData) (int32, error) {
	ver := md.Versions[0]

	//文件数据保存
	if req.Locate == nil {
		var e error
		if ver.Locate, e = streamToDataServer(req, ver.Size); e != nil {
			return -1, e
		}
	} else {
		ver.Locate = req.Locate
	}

	//元数据保存
	return o.metaService.SaveMetadata(md)
}

func streamToDataServer(req *entity.PutReq, size int64) ([]string, error) {
	//stream to store
	stream, e := dataServerStream(req.FileName, size)
	if e != nil {
		return nil, e
	}

	//digest validation
	if pool.Config.EnableHashCheck {
		reader := io.TeeReader(bufio.NewReaderSize(req.Body, 2048), stream)
		hash := util.SHA256Hash(reader)
		if hash != req.Hash {
			log.Infof("Digest of %v validation failure\n", req.Name)
			if e = stream.Commit(false); e != nil {
				log.Errorln(e)
			}
			return nil, ErrInvalidFile
		}
	} else {
		if _, e = io.CopyBuffer(stream, req.Body, make([]byte, 2048)); e != nil {
			if e = stream.Commit(false); e != nil {
				log.Errorln(e)
			}
			return nil, ErrInternalServer
		}
	}

	if e = stream.Commit(true); e != nil {
		log.Errorln(e)
		return nil, ErrServiceUnavailable
	}
	return stream.Locates, e
}

func (o *ObjectService) GetObject(ver *entity.Version) (io.ReadSeekCloser, error) {
	r, e := NewRSGetStream(ver.Size, ver.Hash, ver.Locate)
	if e == ErrNeedUpdateMeta {
		o.metaService.UpdateVersion(ver)
		e = nil
	}
	return r, e
}

func dataServerStream(name string, size int64) (*RSPutStream, error) {
	ds := SelectDataServer(pool.Balancer, pool.Config.Rs.AllShards())
	if len(ds) == 0 {
		return nil, ErrServiceUnavailable
	}
	return NewRSPutStream(ds, name, size)
}
