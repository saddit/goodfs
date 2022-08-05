package app

import (
	"common/registry"
	"fmt"
	"objectserver/config"

	"objectserver/internal/controller/http"
	"objectserver/internal/controller/locate"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/service"
	"os"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func initDir(cfg *config.Config) {
	if !service.ExistPath(cfg.TempPath) {
		if e := os.MkdirAll(cfg.TempPath, os.ModePerm); e != nil {
			panic(e)
		}
	}
	if !service.ExistPath(cfg.StoragePath) {
		if e := os.MkdirAll(cfg.StoragePath, os.ModePerm); e != nil {
			panic(e)
		}
	}
}

func Run(cfg *config.Config) {
	pool.InitPool(cfg)
	defer pool.Close()
	//pre reslove
	initDir(cfg)
	service.WarmUpLocateCache()
	//init components
	httpServer := http.NewHttpServer()
	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints: cfg.Etcd.Endpoint,
		Username:  cfg.Etcd.Username,
		Password:  cfg.Etcd.Password,
	})
	if err != nil {
		logrus.Errorf("create etcd client err: %v", err)
		return
	}
	netAddr := fmt.Sprint(cfg.LocalAddr(), ":", cfg.Port)
	reg := registry.NewEtcdRegistry(etcdCli, cfg.Registry, netAddr)
	// register self
	if err := reg.Register(); err != nil {
		logrus.Errorf("register err: %v", err)
		return
	}
	defer etcdCli.Close()
	defer reg.Unregister()
	//start services
	locate.New(pool.Etcd).StartLocate(netAddr)
	service.StartTempRemovalBackground()
	httpServer.ListenAndServe(netAddr)
}
