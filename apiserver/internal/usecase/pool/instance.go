package pool

import (
	"apiserver/config"
	"apiserver/internal/usecase/selector"
	"common/registry"
	"common/util"
	"net/http"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	Config    *config.Config
	Etcd      *clientv3.Client
	Http      *http.Client
	Balancer  selector.Selector
	Discovery *registry.EtcdDiscovery
)

func InitPool(cfg *config.Config) {
	Config = cfg
	initHttpClient()
	initEtcd(cfg)
	//initDiscovery(Etcd, &cfg.Registry)
	initBalancer(cfg)
}

func Close() {
	Http.CloseIdleConnections()
	util.LogErr(Etcd.Close())
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

func initDiscovery(etcd *clientv3.Client, cfg *registry.Config) {
	Discovery = registry.NewEtcdDiscovery(etcd, cfg)
}

func initHttpClient() {
	Http = &http.Client{Timeout: 5 * time.Second}
}

func initBalancer(cfg *config.Config) {
	Balancer = selector.NewSelector(cfg.SelectStrategy)
}
