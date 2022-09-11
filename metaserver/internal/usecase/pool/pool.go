package pool

import (
	"common/etcd"
	"common/hashslot"
	"common/logs"
	"common/util"
	"context"
	"fmt"
	"metaserver/config"
	"metaserver/internal/usecase/db"
	"metaserver/internal/usecase/raftimpl"
	"sort"
	"strings"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	Config       *config.Config
	Storage      *db.Storage
	RaftWrapper  *raftimpl.RaftWrapper
	Etcd         *clientv3.Client
	HashSlots    hashslot.IEdgeProvider
	HttpHostPort string
	GrpcHostPort string
)

func InitPool(cfg *config.Config) {
	Config = cfg
	HttpHostPort = util.GetHostPort(cfg.Port)
	GrpcHostPort = util.GetHostPort(cfg.Cluster.Port)
	initEtcd(&cfg.Etcd)
	initStorage(cfg)
	initHashSlot(cfg, Etcd)
}

func initEtcd(cfg *etcd.Config) {
	var err error
	Etcd, err = clientv3.New(clientv3.Config{
		Endpoints: cfg.Endpoint,
		Username:  cfg.Username,
		Password:  cfg.Password,
	})
	if err != nil {
		panic(fmt.Errorf("create etcd client err: %v", err))
	}
}

func initStorage(cfg *config.Config) {
	// open db file
	Storage = db.NewStorage()
	if err := Storage.Open(cfg.DataDir); err != nil {
		panic(fmt.Errorf("open db err: %v", err))
	}
}

func initHashSlot(cfg *config.Config, etcd *clientv3.Client) {
	var err error
	// sort slot to keep incr order
	sort.Strings(cfg.HashSlot)
	slotStr := strings.Join(cfg.HashSlot, ",")
	logs.Std().Infof("hash slots: %s", slotStr)
	// save current slot data
	resp, err := etcd.Put(context.Background(), fmt.Sprint("metaserver_hashslot/", HttpHostPort), slotStr, clientv3.WithPrevKV())
	if err != nil {
		panic(fmt.Errorf("save hash slot to etcd err: %s", err))
	}
	// get slots data from etcd
	res, err := etcd.Get(context.Background(), "metaserver_hashslot", clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}
	// wrap slots
	slotsMap := make(map[string][]string)
	for _, kv := range res.Kvs {
		identify := strings.Split(string(kv.Key), "/")[1]
		slots := strings.Split(string(kv.Value), ",")
		slotsMap[identify] = slots
	}
	HashSlots, err = hashslot.WrapSlots(slotsMap)
	if err != nil {
		panic(fmt.Errorf("init hash slot err: %s", err))
	}
	// if prevKv exists and doesn't equal to current setting
	if resp.PrevKv != nil && string(resp.PrevKv.Value) != slotStr {
		// TODO migration data through a raft way
		//  1. find out removed slots
		//  2. delete and transfer to another server
		//   2.1 map all kvs on bolt-db, recheck the mapping from key to hash-slot by newest hash-slots
		//   2.2 delete fail: log error and skip
		//   2.3 transfer fail: rollback and log error
		//  3. if some data migrate failure, log and panic.
		//     administrator should find out reason and retry.
	}
}

func Close() {
	util.LogErr(Storage.Stop())
	util.LogErr(Etcd.Close())
	if RaftWrapper != nil {
		util.LogErr(RaftWrapper.Close())
	}
}
