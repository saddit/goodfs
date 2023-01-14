package app

import (
	"common/graceful"
	"common/logs"
	"common/util"
	"objectserver/config"
	"objectserver/internal/controller/grpc"
	"objectserver/internal/controller/http"
	"objectserver/internal/controller/locate"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/service"
)

func Run(cfg *config.Config) {
	// init logger
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
	// driver manger
	driverManager := service.NewDriverManager(service.NewFreeFirstDriver())
	defer driverManager.StartAutoUpdate()()
	// warmup serv
	service.WarmUpLocateCache()
	// startup server
	graceful.ListenAndServe(http.NewHttpServer(netAddr), grpc.NewRpcServer(cfg.RpcPort, service.NewMigrationService(pool.ObjectCap)))
}
