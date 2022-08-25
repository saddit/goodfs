package pool

import (
	"common/etcd"
	"common/util"
	"fmt"
	"metaserver/config"
	"metaserver/internal/usecase/db"
	"metaserver/internal/usecase/raftimpl"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	Config      *config.Config
	Storage     *db.Storage
	RaftWrapper *raftimpl.RaftWrapper
	Etcd        *clientv3.Client
)

func InitPool(cfg *config.Config) {
	initEtcd(&cfg.Etcd)
	initStorage(cfg)
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

func Close() {
	util.LogErr(Storage.Close())
	util.LogErr(Etcd.Close())
	if RaftWrapper != nil {
		util.LogErr(RaftWrapper.Close())
	}
}
