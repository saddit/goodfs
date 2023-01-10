package pool

import (
	"adminserver/config"
	"adminserver/internal/usecase/db"
	"common/registry"
	"common/util"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net/http"
)

var (
	Config    *config.Config
	Etcd      *clientv3.Client
	Http      *http.Client
	Discovery *registry.EtcdDiscovery
	StatDB    *db.ServerStatDB
)

func Init(cfg *config.Config) {
	Config = cfg
	initHttpClient()
	initEtcd(cfg)
	initDiscovery(Etcd, cfg)
	initStatDB(Etcd, cfg)
}

func Close() {
	Http.CloseIdleConnections()
	util.LogErr(Etcd.Close())
	util.LogErr(StatDB.Close())
}

func initHttpClient() {
	Http = &http.Client{Timeout: 0}
}

func initEtcd(cfg *config.Config) {
	// init etcd
	var err error
	Etcd, err = clientv3.New(clientv3.Config{
		Endpoints: cfg.Etcd.Endpoint,
		Username:  cfg.Etcd.Username,
		Password:  cfg.Etcd.Password,
	})
	if err != nil {
		panic(err)
	}
}

func initDiscovery(etcd *clientv3.Client, cfg *config.Config) {
	conf := cfg.GetRegistryCfg()
	conf.Services = []string{cfg.Discovery.DataServName, cfg.Discovery.MetaServName, cfg.Discovery.ApiServName}
	Discovery = registry.NewEtcdDiscovery(etcd, conf)
}

func initStatDB(etcd *clientv3.Client, cfg *config.Config) {
	services := []string{cfg.Discovery.DataServName, cfg.Discovery.MetaServName, cfg.Discovery.ApiServName}
	cli := db.ServerStatCli{Watcher: etcd, KV: etcd}
	StatDB = db.NewServerStatDB(cli, cfg.Discovery.Group, services)
}
