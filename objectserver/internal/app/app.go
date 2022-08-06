package app

import (
	"common/logs"
	"common/registry"
	"common/util"
	"fmt"
	"objectserver/config"

	"objectserver/internal/controller/http"
	"objectserver/internal/controller/locate"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/service"
	"os"
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
	initDir(cfg)
	netAddr := fmt.Sprint(util.GetHost(), ":", cfg.Port)
	logs.SetLevel(cfg.LogLevel)
	//init components
	pool.InitPool(cfg)
	defer pool.Close()
	// register
	defer registry.NewEtcdRegistry(pool.Etcd, cfg.Registry, netAddr).MustRegister().Unregister()
	// locating serv
	defer locate.New(pool.Etcd).StartLocate(netAddr)()
	// cleaning serv
	defer service.StartTempRemovalBackground(pool.Cache)()
	// warmup serv
	service.WarmUpLocateCache()
	http.NewHttpServer().ListenAndServe(netAddr)
}
