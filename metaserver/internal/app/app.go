package app

import (
	"common/graceful"
	"common/logs"
	"common/registry"
	"common/util"
	. "metaserver/config"
	"metaserver/internal/controller/grpc"
	"metaserver/internal/controller/http"
	"metaserver/internal/usecase/repo"
	"metaserver/internal/usecase/service"
	"metaserver/internal/usecase/pool"
)

func Run(cfg *Config) {
	// init logger
	logs.SetLevel(cfg.LogLevel)
	// init components
	pool.InitPool(cfg)
	defer pool.Close()
	// init services
	var grpcServer *grpc.RpcRaftServer
	metaRepo := repo.NewMetadataRepo(pool.Storage)
	metaService := service.NewMetadataService(metaRepo)
	grpcServer, pool.RaftWrapper = grpc.NewRpcRaftServer(cfg.Cluster, metaRepo)
	httpServer := http.NewHttpServer(pool.HttpHostPort, metaService)
	reg := registry.NewEtcdRegistry(pool.Etcd, cfg.Registry, pool.HttpHostPort)
	// register on leader change
	pool.RaftWrapper.OnLeaderChanged = func(isLeader bool) {
		util.LogErr(reg.Unregister())
		if isLeader {
			util.LogErr(reg.AsMaster().Register())
		} else {
			util.LogErr(reg.AsSlave().Register())
		}
	}
	// register first time
	if pool.RaftWrapper.Enabled {
		reg.AsSlave()
	}
	defer reg.MustRegister().Unregister()
	graceful.ListenAndServe(httpServer, grpcServer)
}
