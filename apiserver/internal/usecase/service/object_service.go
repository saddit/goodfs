package service

import (
	"apiserver/internal/entity"
	. "apiserver/internal/usecase"
	"apiserver/internal/usecase/pool"
	"bufio"
	"common/util"
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type ObjectService struct {
	metaService IMetaService
	etcd        *clientv3.Client
}

func NewObjectService(s IMetaService, etcd *clientv3.Client) *ObjectService {
	return &ObjectService{s, etcd}
}

// LocateObject 根据Hash定位所有分片位置
func (o *ObjectService) LocateObject(hash string) ([]string, bool) {
	//利用etcd监听机制实现
	// 1. 临时建立并watch一个唯一key
	// 2. object_server 持续watch各自唯一key
	// 3. set每个object_server的key为（hash+临时key）
	// 4. objectserver watch到这个变化就定位，成功则向set临时key
	// 5. api服务器watch得到文件位置
	tempId := uuid.NewString()
	ctx := context.Background()
	if _, err := o.etcd.Put(ctx, tempId, ""); err != nil {
		logrus.Error(err)
		return nil, false
	}
	defer o.etcd.Delete(ctx, tempId)
	wt := o.etcd.Watch(ctx, tempId)
	locates := make([]string, 0, pool.Config.Rs.AllShards())
	//TODO 从discovery中获取所有objectserver的地址，并SET他们监听的key为 hash+tempId
	for {
		select {
		case resp, ok := <-wt:
			if !ok {
				logrus.Error("Etcd watching key err, channel closed")
				return nil, false
			}
			for _, event := range resp.Events {
				locates = append(locates, string(event.Kv.Value))
			}
			if len(locates) == cap(locates) {
				return locates, true
			}
		case <-time.Tick(time.Minute):
			logrus.Errorf("locate object %s timeout!", hash)
			return nil, false
		}
	}
}

func (o *ObjectService) StoreObject(req *entity.PutReq, md *entity.Metadata) (int32, error) {
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
