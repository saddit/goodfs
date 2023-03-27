package app

import (
	"common/cst"
	"common/graceful"
	"common/registry"
	"common/system"
	clientv3 "go.etcd.io/etcd/client/v3"
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
	// init components
	lifecycle := registry.NewLifecycle(pool.Etcd, cfg.Registry.Interval)
	syncer := system.Syncer(pool.Etcd, cst.EtcdPrefix.FmtSystemInfo(cfg.Registry.Group, cfg.Registry.Name, cfg.Registry.SID()))
	// lifecycle event
	lifecycle.Subscribe(pool.Registry.Register)
	lifecycle.Subscribe(func(id clientv3.LeaseID) {
		syncer.LeaseID = id
		_ = syncer.Sync()
	})
	pool.OnOpen(func() {
		go lifecycle.DeadLoop()
		pool.OnClose(
			func() { lifecycle.Close() },
			// unregister
			func() { pool.Registry.Unregister() },
			// locating serv
			service.NewLocator(pool.Etcd).StartLocate(),
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
