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
	"fmt"
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
	// register
	cfg.Registry.ServerPort = cfg.Port
	reg := registry.NewEtcdRegistry(pool.Etcd, cfg.Registry)
	defer reg.MustRegister().Unregister()
	// system-info auto saving
	syncer := system.Syncer(pool.Etcd, fmt.Sprint(cst.EtcdPrefix.SystemInfo, "/", pool.Config.Registry.RegisterKey()), <-reg.LifecycleLease())
	defer syncer.StartAutoSave()()
	//start api server
	graceful.ListenAndServe(
		nil,
		http.NewHttpServer(cfg.Port, objService, metaService, bucketRepo),
	)
}
