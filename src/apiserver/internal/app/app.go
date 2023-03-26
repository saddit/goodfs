package app

import (
	. "apiserver/config"
	"apiserver/internal/controller/http"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/repo"
	"apiserver/internal/usecase/service"
	"common/cst"
	"common/graceful"
	"common/registry"
	"common/system"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func Run(cfg *Config) {
	pool.InitPool(cfg)
	defer pool.Close()
	//init services
	versionRepo := repo.NewVersionRepo()
	bucketRepo := repo.NewBucketRepo()
	metaRepo := repo.NewMetadataRepo()
	metaService := service.NewMetaService(metaRepo, versionRepo)
	objService := service.NewObjectService(metaService, bucketRepo, pool.Etcd)

	// lifecycle
	lifecycle := registry.NewLifecycle(pool.Etcd, cfg.Registry.Interval)
	defer lifecycle.Close()

	// register
	reg := registry.NewEtcdRegistry(pool.Etcd, &cfg.Registry)
	lifecycle.Subscribe(reg.Register)
	defer reg.Unregister()

	// system-info auto saving
	syncer := system.Syncer(pool.Etcd, cst.EtcdPrefix.FmtSystemInfo(cfg.Registry.Group, cfg.Registry.Name, cfg.Registry.SID()))
	lifecycle.Subscribe(func(id clientv3.LeaseID) {
		syncer.LeaseID = id
		_ = syncer.Sync()
	})
	defer syncer.StartAutoSave()()

	// start lifecycle
	go lifecycle.DeadLoop()

	//start api server
	graceful.ListenAndServe(
		nil,
		http.NewHttpServer(cfg.Port, objService, metaService, bucketRepo),
	)
}
