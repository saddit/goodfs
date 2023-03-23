package app

import (
	"common/cst"
	"common/graceful"
	"common/system"
	"common/util"
	"objectserver/config"
	"objectserver/internal/controller/grpc"
	"objectserver/internal/controller/http"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/service"
)

func Run(cfg *config.Config) {
	//init components
	pool.InitPool(cfg)
	defer pool.CloseAll()
	netAddr := util.ServerAddress(cfg.Port)
	pool.OnOpen(func() {
		// register service
		pool.Registry.MustRegister()
		syncer := system.Syncer(pool.Etcd, cst.EtcdPrefix.FmtSystemInfo(cfg.Registry.Group, cfg.Registry.Name, cfg.Registry.SID()), <-pool.Registry.LifecycleLease())
		pool.OnClose(
			// unregister
			func() { pool.Registry.Unregister() },
			// locating serv
			service.NewLocator(pool.Etcd).StartLocate(netAddr),
			// cleaning serv
			service.StartTempRemovalBackground(pool.Cache, pool.Config.TempCleaners),
			// auto update driver stat
			pool.DriverManager.StartAutoUpdate(),
			// system info sync
			syncer.StartAutoSave(),
		)
	})
	pool.Open()
	// warmup serv
	service.WarmUpLocateCache()
	// startup server
	grpcServer := grpc.NewServer(service.NewMigrationService(pool.ObjectCap))
	graceful.ListenAndServe(nil, http.NewHttpServer(cfg.Port, grpcServer))
}
