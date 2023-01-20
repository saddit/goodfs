package app

import (
	"common/graceful"
	"common/util"
	"objectserver/config"
	"objectserver/internal/controller/grpc"
	"objectserver/internal/controller/http"
	"objectserver/internal/controller/locate"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/service"
)

func Run(cfg *config.Config) {
	//init components
	pool.InitPool(cfg)
	defer pool.Etcd.Close()
	defer pool.Cache.Close()
	defer pool.Close()
	netAddr := util.GetHostPort(cfg.Port)
	pool.OnOpen(func() {
		// register service
		util.PanicErr(pool.Registry.Register())
		pool.OnClose(
			// unregister
			func() { pool.Registry.Unregister() },
			// locating serv
			locate.New(pool.Etcd).StartLocate(netAddr),
			// cleaning serv
			service.StartTempRemovalBackground(pool.Cache),
			// auto save server capacity info
			pool.ObjectCap.StartAutoSave(cfg.State.SyncInterval),
		)
	})
	pool.Open()
	// warmup serv
	service.WarmUpLocateCache()
	// startup server
	graceful.ListenAndServe(nil, http.NewHttpServer(netAddr), grpc.NewRpcServer(cfg.RpcPort, service.NewMigrationService(pool.ObjectCap)))
}
