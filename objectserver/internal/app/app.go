package app

import (
	"common/graceful"
	"common/logs"
	"common/registry"
	"common/util"
	"objectserver/config"
	"objectserver/internal/controller/grpc"

	"objectserver/internal/controller/http"
	"objectserver/internal/controller/locate"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/service"
	"os"
)

func initDir(cfg *config.Config) {
	if e := os.MkdirAll(cfg.TempPath, util.OS_ModeUser); e != nil {
		panic(e)
	}
	if e := os.MkdirAll(cfg.StoragePath, util.OS_ModeUser); e != nil {
		panic(e)
	}
}

func Run(cfg *config.Config) {
	initDir(cfg)
	logs.SetLevel(cfg.LogLevel)
	//init components
	pool.InitPool(cfg)
	defer pool.Close()
	netAddr := util.GetHostPort(cfg.Port)
	// register
	defer registry.NewEtcdRegistry(pool.Etcd, cfg.Registry, netAddr).MustRegister().Unregister()
	// locating serv
	defer locate.New(pool.Etcd).StartLocate(netAddr)()
	// cleaning serv
	defer service.StartTempRemovalBackground(pool.Cache)()
	// warmup serv
	service.WarmUpLocateCache()
	// startup server
	graceful.ListenAndServe(http.NewHttpServer(netAddr), grpc.NewRpcServer(cfg.RpcPort, new(service.MigrationService)))
}
