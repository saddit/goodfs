package app

import (
	"common/cst"
	"common/graceful"
	"common/logs"
	"common/util"
	"objectserver/config"
	"objectserver/internal/controller/grpc"
	"os"

	"objectserver/internal/controller/http"
	"objectserver/internal/controller/locate"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/service"
)

func initDir(cfg *config.Config) {
	if e := os.MkdirAll(cfg.TempPath, cst.OS.ModeUser); e != nil {
		panic(e)
	}
	if e := os.MkdirAll(cfg.StoragePath, cst.OS.ModeUser); e != nil {
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
	// register service
	util.PanicErr(pool.Registry.Register())
	defer pool.Registry.Unregister()
	// locating serv
	defer locate.New(pool.Etcd).StartLocate(netAddr)()
	// cleaning serv
	defer service.StartTempRemovalBackground(pool.Cache)()
	// auto save server capacity info
	defer pool.ObjectCap.StartAutoSave(cfg.State.SyncInterval)()
	// warmup serv
	service.WarmUpLocateCache()
	// startup server
	graceful.ListenAndServe(http.NewHttpServer(netAddr), grpc.NewRpcServer(cfg.RpcPort, service.NewMigrationService(pool.ObjectCap)))
}
