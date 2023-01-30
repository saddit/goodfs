package service

import (
	"apiserver/internal/entity"
	. "apiserver/internal/usecase"
	"apiserver/internal/usecase/logic"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/repo"
	"bufio"
	"common/cst"
	"common/graceful"
	"common/logs"
	"common/response"
	"common/system"
	"common/util"
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

// getLocateResp raw must like "ip#idx"
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
	bucketRepo  repo.IBucketRepo
	etcd        *clientv3.Client
	sysStat     map[string]*system.Info
}

func NewObjectService(s IMetaService, b repo.IBucketRepo, etcd *clientv3.Client) *ObjectService {
	return &ObjectService{s, b, etcd, make(map[string]*system.Info)}
}

// LocateObject locate object shards by hash. send "hash.idx#key" expect "ip#idx"
func (o *ObjectService) LocateObject(hash string) ([]string, bool) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// generate a unique id as key for receive locates
	tempId := uuid.NewString()
	if _, err := o.etcd.Put(ctx, tempId, ""); err != nil {
		logs.Std().Error(err)
		return nil, false
	}
	// remove this key after all
	defer o.etcd.Delete(ctx, tempId)
	wt := o.etcd.Watch(ctx, tempId)
	locates := make([]string, pool.Config.Object.ReedSolomon.AllShards())
	for i := 0; i < pool.Config.Object.ReedSolomon.AllShards(); i++ {
		val := fmt.Sprintf("%s.%d#%s", hash, i, tempId)
		_, err := o.etcd.Put(ctx, cst.EtcdPrefix.LocationSubKey, val)
		if err != nil {
			logs.Std().Errorf("put '%s' to location-sub-key err: %s", val, err)
		}
	}
	// to receive locates
	tt := time.NewTicker(pool.Config.LocateTimeout)
	defer tt.Stop()
	var cnt int
	for {
		select {
		case resp, ok := <-wt:
			if !ok {
				logs.Std().Error("etcd watching locate-key err, channel closed")
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

// StoreObject store object to data server
func (o *ObjectService) StoreObject(req *entity.PutReq, md *entity.Metadata) (vn int32, err error) {
	dg := util.NewDoneGroup()
	defer dg.Close()
	// get metadata if exist
	var metadata *entity.Metadata
	dg.Todo()
	go func() {
		defer dg.Done()
		var inner error
		metadata, inner = o.metaService.GetMetadata(md.Name, md.Bucket, int32(entity.VerModeNot), true)
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
		bucket, inner = o.bucketRepo.Get(md.Bucket)
		if inner != nil {
			dg.Error(inner)
		}
	}()
	// wait
	if err = dg.WaitUntilError(); err != nil {
		return
	}
	// check bucket writable
	if bucket.Readonly {
		return 0, response.NewError(400, "bucket is readonly")
	}

	ver := md.Versions[0]
	FillVersionConfig(ver, bucket)
	provider := NewStreamProvider(&StreamOption{
		Bucket:   bucket.Name,
		Hash:     ver.Hash,
		Name:     metadata.Name,
		Size:     ver.Size,
		Compress: ver.Compress,
	}, ver)
	// if object not exists, upload to data server
	if len(req.Locate) == 0 {
		if ver.Locate, err = streamToDataServer(req, ver, provider); err != nil {
			return -1, fmt.Errorf("stream to data server err: %w", err)
		}
	} else {
		// otherwise save locates
		ver.Locate = req.Locate
	}
	// save metadata
	if metadata == nil {
		return o.metaService.SaveMetadata(md)
	}
	if vn, err = o.metaService.AddVersion(md.Name, md.Bucket, ver); err != nil {
		return
	}
	if metadata.Total > 0 && !bucket.Versioning || metadata.Total >= bucket.VersionRemains {
		go func() {
			defer graceful.Recover()
			// if not err, delete first version
			inner := o.metaService.RemoveVersion(md.Name, md.Bucket, int32(metadata.FirstVersion))
			util.LogErrWithPre("remove first version err", inner)
		}()
	}
	return
}

func streamToDataServer(req *entity.PutReq, meta *entity.Version, provider StreamProvider) ([]string, error) {
	//stream to store
	stream, locates, err := dataServerStream(meta, provider)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	//digest validation
	if pool.Config.Checksum {
		reader := io.TeeReader(bufio.NewReaderSize(req.Body, 8*cst.OS.PageSize), stream)
		hash := crypto.SHA256IO(reader)
		if hash != req.Hash {
			logs.Std().Infof("Digest of %v validation failure\n", req.Name)
			if err = stream.Commit(false); err != nil {
				logs.Std().Errorln(err)
			}
			return nil, ErrInvalidFile
		}
	} else {
		// copy request body to stream
		if _, err = io.CopyBuffer(stream, req.Body, make([]byte, 8*cst.OS.PageSize)); err != nil {
			logs.Std().Error(err)
			if err = stream.Commit(false); err != nil {
				logs.Std().Errorln(err)
			}
			return nil, ErrInternalServer
		}
	}
	// upload success
	if err = stream.Commit(true); err != nil {
		logs.Std().Errorln(err)
		return nil, ErrServiceUnavailable
	}
	return locates, nil
}

func (o *ObjectService) GetObject(meta *entity.Metadata, ver *entity.Version) (io.ReadSeekCloser, error) {
	up := func(locates []string) error {
		ver.Locate = locates
		return o.metaService.UpdateVersion(meta.Name, meta.Bucket, ver)
	}
	opt := &StreamOption{
		Hash:     ver.Hash,
		Size:     ver.Size,
		Name:     meta.Name,
		Bucket:   meta.Bucket,
		Compress: ver.Compress,
		Updater:  up,
	}
	return NewStreamProvider(opt, ver).GetStream(ver.Locate)
}

func (o *ObjectService) GetServerSystemInfo(serverId string) *system.Info {
	return o.sysStat[serverId]
}

func (o *ObjectService) WatchingObjectServerStat() func() {
	keyPrefix := cst.EtcdPrefix.FmtSystemInfo(pool.Config.Registry.Group, pool.Config.Discovery.DataServName, "")
	// get exists system stat first
	resp, err := o.etcd.Get(context.Background(), keyPrefix, clientv3.WithPrefix())
	if err == nil {
		for _, kv := range resp.Kvs {
			serverId := util.BytesToStr(kv.Key)
			var stat system.Info
			if err := util.DecodeMsgp(&stat, kv.Value); err != nil {
				logs.Std().Errorf("decode System.Info err: %s", err)
				continue
			}
			o.sysStat[serverId] = &stat
		}
	} else {
		logs.Std().Errorf("get exists object-server sys-info fail: %s", err)
	}
	// start watching prefix
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer graceful.Recover()
		// retry if abort by error
		for {
			var success bool
			watchChan := o.etcd.Watch(ctx, keyPrefix, clientv3.WithPrefix(), clientv3.WithFilterDelete())
			for events := range watchChan {
				if events.Canceled {
					logs.Std().Errorf("watching object-server system stat abort: %s", events.Err())
					success = false
					break
				}
				for _, event := range events.Events {
					serverId := util.BytesToStr(event.Kv.Key)
					var stat system.Info
					if err := util.DecodeMsgp(&stat, event.Kv.Value); err != nil {
						logs.Std().Errorf("decode System.Info err: %s", err)
						continue
					}
					o.sysStat[serverId] = &stat
				}
			}
			// quit if canceled by context
			if success {
				break
			}
			// sleep 2 sec before retry
			time.Sleep(2 * time.Second)
		}
	}()
	return cancel
}

func NewStreamProvider(opt *StreamOption, ver *entity.Version) StreamProvider {
	switch ver.StoreStrategy {
	default:
		fallthrough
	case entity.ECReedSolomon:
		cfg := pool.Config.Object.ReedSolomon
		cfg.DataShards = ver.DataShards
		cfg.ParityShards = ver.ParityShards
		return RsStreamProvider(opt, &cfg)
	case entity.MultiReplication:
		cfg := pool.Config.Object.Replication
		cfg.CopiesCount = ver.DataShards
		return CpStreamProvider(opt, &cfg)
	}
}

func FillVersionConfig(ver *entity.Version, bucket *entity.Bucket) {
	if bucket.Compress {
		ver.Compress = true
	}
	rsConf := pool.Config.Object.ReedSolomon
	rpConf := pool.Config.Object.Replication

	if bucket.StoreStrategy > 0 {
		ver.StoreStrategy = bucket.StoreStrategy
		rsConf.DataShards = bucket.DataShards
		rsConf.ParityShards = bucket.ParityShards
		rpConf.CopiesCount = bucket.DataShards
	}

	switch ver.StoreStrategy {
	default:
		fallthrough
	case entity.ECReedSolomon:
		ver.DataShards = rsConf.DataShards
		ver.ParityShards = rsConf.ParityShards
		ver.ShardSize = rsConf.ShardSize(ver.Size)
	case entity.MultiReplication:
		ver.DataShards = rpConf.CopiesCount
		ver.ShardSize = int(ver.Size)
	}
}

func dataServerStream(meta *entity.Version, provider StreamProvider) (WriteCommitCloser, []string, error) {
	ds := logic.NewDiscovery().SelectDataServer(pool.Balancer, meta.DataShards+meta.ParityShards)
	if len(ds) == 0 {
		return nil, nil, ErrServiceUnavailable
	}

	stream, err := provider.PutStream(ds)
	return stream, ds, err
}
