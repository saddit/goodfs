package app

import (
	. "apiserver/config"
	"apiserver/internal/controller/http"
	"apiserver/internal/usecase/logic"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/repo"
	"apiserver/internal/usecase/service"
	"common/graceful"
	"common/registry"
	"common/util"
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
	cfg.Registry.HttpAddr = util.ServerAddress(cfg.Port)
	defer registry.NewEtcdRegistry(pool.Etcd, cfg.Registry).MustRegister().Unregister()
	// system-info auto saving
	defer logic.NewSystemStatLogic().StartAutoSave()()
	//start api server
	graceful.ListenAndServe(
		nil,
		http.NewHttpServer(cfg.Port, objService, metaService, bucketRepo),
	)
}
