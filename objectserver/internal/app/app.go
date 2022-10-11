package app

import (
	"common/graceful"
	"common/logs"
	"common/util"
	"objectserver/internal/usecase/logic"
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
	util.PanicErr(pool.Registry.Register())
	defer pool.Registry.Unregister()
	// locating serv
	defer locate.New(pool.Etcd).StartLocate(netAddr)()
	// cleaning serv
	defer service.StartTempRemovalBackground(pool.Cache)()
	util.PanicErr(logic.NewPeers().RegisterSelf())
	defer logic.NewPeers().UnregisterSelf()
	// warmup serv
	service.WarmUpLocateCache()
	// startup server
	graceful.ListenAndServe(http.NewHttpServer(netAddr), grpc.NewRpcServer(cfg.RpcPort, service.NewMigrationService(pool.ObjectCap)))
}
