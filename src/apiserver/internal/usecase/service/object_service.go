package service

import (
	"apiserver/internal/entity"
	. "apiserver/internal/usecase"
	"apiserver/internal/usecase/logic"
	"apiserver/internal/usecase/pool"
	"bufio"
	"common/logs"
	"common/util/crypto"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	LocationSubKey = "goodfs.location"
)

// getLocateResp must like "ip#idx"
func getLocateResp(raw string) (ip string, idx int) {
	var err error
	strs := strings.Split(raw, "#")
	if len(strs) != 2 {
		panic("err format locating resp: " + raw)
	}
	idx, err = strconv.Atoi(strs[1])
	if err != nil {
		panic(err)
	}
	ip = strs[0]
	return
}

type ObjectService struct {
	metaService IMetaService
	etcd        *clientv3.Client
}

func NewObjectService(s IMetaService, etcd *clientv3.Client) *ObjectService {
	return &ObjectService{s, etcd}
}

// LocateObject locate object shards by hash. send "hash.idx#key" expect "ip#idx"
func (o *ObjectService) LocateObject(hash string) ([]string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	//生成一个唯一key 并在结束后删除
	tempId := uuid.NewString()
	if _, err := o.etcd.Put(ctx, tempId, ""); err != nil {
		logs.Std().Error(err)
		return nil, false
	}
	defer o.etcd.Delete(ctx, tempId)
	wt := o.etcd.Watch(ctx, tempId)
	locates := make([]string, pool.Config.Rs.AllShards())
	for i := 0; i < pool.Config.Rs.AllShards(); i++ {
		o.etcd.Put(ctx, LocationSubKey, fmt.Sprintf("%s.%d#%s", hash, i, tempId))
	}
	//开始监听变化
	tt := time.NewTicker(pool.Config.LocateTimeout)
	defer tt.Stop()
	var cnt int
	for {
		select {
		case resp, ok := <-wt:
			if !ok {
				logs.Std().Error("etcd watching key err, channel closed")
				return nil, false
			}
			for _, event := range resp.Events {
				ip, idx := getLocateResp(string(event.Kv.Value))
				logs.Std().Debugf("located success for index %d of %s at %s", idx, hash, ip)
				locates[idx] = ip
				cnt += 1
			}
			if len(locates) == cnt {
				return locates, true
			}
		case <-tt.C:
			logs.Std().Warnf("locate object %s timeout!", hash)
			return nil, false
		}
	}
}

func (o *ObjectService) StoreObject(req *entity.PutReq, md *entity.Metadata) (int32, error) {
	ver := md.Versions[0]

	//文件数据保存
	if len(req.Locate) == 0 {
		var e error
		if ver.Locate, e = streamToDataServer(req, ver); e != nil {
			return -1, e
		}
	} else {
		ver.Locate = req.Locate
	}
	//元数据保存
	return o.metaService.SaveMetadata(md)
}

func streamToDataServer(req *entity.PutReq, meta *entity.Version) ([]string, error) {
	//stream to store
	stream, locates, e := dataServerStream(meta)
	if e != nil {
		return nil, e
	}
	defer stream.Close()

	//digest validation
	if pool.Config.Checksum {
		reader := io.TeeReader(bufio.NewReaderSize(req.Body, 32 * 1024), stream)
		hash := crypto.SHA256IO(reader)
		if hash != req.Hash {
			logs.Std().Infof("Digest of %v validation failure\n", req.Name)
			if e = stream.Commit(false); e != nil {
				logs.Std().Errorln(e)
			}
			return nil, ErrInvalidFile
		}
	} else {
		if _, e = io.CopyBuffer(stream, req.Body, make([]byte, 32 * 1024)); e != nil {
			if e = stream.Commit(false); e != nil {
				logs.Std().Errorln(e)
			}
			return nil, ErrInternalServer
		}
	}

	if e = stream.Commit(true); e != nil {
		logs.Std().Errorln(e)
		return nil, ErrServiceUnavailable
	}
	return locates, e
}

func (o *ObjectService) GetObject(meta *entity.Metadata, ver *entity.Version) (io.ReadSeekCloser, error) {
	var (
		stream io.ReadSeekCloser
		err error
	)

	switch ver.StoreStrategy {
	default:
		fallthrough
	case entity.ECReedSolomon:
		stream, err = NewRSGetStream(ver.Size, ver.Hash, ver.Locate, &pool.Config.Rs)
	case entity.MultiReplication:
		stream, err = NewCopyGetStream(ver.Hash, ver.Locate, ver.Size, &pool.Config.Object.Replication)
	}
	
	if err == ErrNeedUpdateMeta {
		logs.Std().Debugf("data fix: need update meta %s verison %d", meta.Name, ver.Sequence)
		err = o.metaService.UpdateVersion(meta.Name, ver)
	}
	return stream, err
}

func dataServerStream(meta *entity.Version) (WriteCloseCommitter, []string, error) {
	var dsNum int
	switch meta.StoreStrategy {
	default:
		fallthrough
	case entity.ECReedSolomon:
		dsNum = pool.Config.Object.ReedSolomon.DataShards + pool.Config.Object.ReedSolomon.ParityShards
	case entity.MultiReplication:
		dsNum = pool.Config.Object.Replication.CopiesCount
	}

	ds := logic.NewDiscovery().SelectDataServer(pool.Balancer, dsNum)
	if len(ds) == 0 {
		return nil, nil, ErrServiceUnavailable
	}
	
	var (
		stream WriteCloseCommitter
		err error
	)
	//兼容不同对象保存策略
	switch meta.StoreStrategy {
	default:
		fallthrough
	case entity.ECReedSolomon:
		cfg := &pool.Config.Object.ReedSolomon
		meta.DataShards = cfg.DataShards
		meta.ParityShards = cfg.ParityShards
		meta.ShardSize = cfg.BlockPerShard
		stream, err = NewRSPutStream(ds, meta.Hash, meta.Size, cfg)
	case entity.MultiReplication:
		cfg := &pool.Config.Object.Replication
		meta.DataShards = cfg.CopiesCount
		meta.ShardSize = int(meta.Size)
		stream, err = NewCopyPutStream(meta.Hash, meta.Size, ds, cfg)
	}
	return stream, ds, err
}
