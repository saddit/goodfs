package app

import (
	. "apiserver/config"
	"apiserver/internal/controller/http"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/repo"
	"apiserver/internal/usecase/service"
	"common/logs"
	"common/registry"
	"common/util"
	"fmt"
)

func Run(cfg *Config) {
	pool.InitPool(cfg)
	defer pool.Close()
	// init log
	logs.SetLevel(cfg.LogLevel)
	//init services
	netAddr := fmt.Sprint(util.GetHost(), ":", cfg.Port)
	versionRepo := repo.NewVersionRepo(pool.Etcd)
	metaRepo := repo.NewMetadataRepo(pool.Etcd, versionRepo)
	metaService := service.NewMetaService(metaRepo, versionRepo)
	objService := service.NewObjectService(metaService, pool.Etcd)
	// register
	defer registry.NewEtcdRegistry(pool.Etcd, cfg.Registry, netAddr).MustRegister().Unregister()
	//start api server
	apiServer := http.NewHttpServer(objService, metaService)
	apiServer.ListenAndServe(netAddr)
}
