package app

import (
	"common/graceful"
	"common/logs"
	. "metaserver/config"
	"metaserver/internal/controller/grpc"
	"metaserver/internal/controller/http"
	"metaserver/internal/usecase/logic"
	"metaserver/internal/usecase/pool"
	"metaserver/internal/usecase/repo"
	"metaserver/internal/usecase/service"
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
	hsService := service.NewHashSlotService(pool.HashSlot, metaRepo, &cfg.HashSlot)
	grpcServer, pool.RaftWrapper = grpc.NewRpcRaftServer(cfg.Cluster, metaRepo, hsService)
	httpServer := http.NewHttpServer(pool.HttpHostPort, metaService)
	// register on leader change
	pool.RaftWrapper.RegisterLeaderChangedEvent(hsService)
	pool.RaftWrapper.RegisterLeaderChangedEvent(logic.NewRegistry())
	// register first time
	if pool.RaftWrapper.Enabled {
		pool.Registry.AsSlave()
		// register peers
		peersLogic := logic.NewPeers()
		if err := peersLogic.RegisterSelf(); err != nil {
			logs.Std().Error(err)
			return
		}
		defer peersLogic.UnregisterSelf()
	} else {
		hsService.OnLeaderChanged(true)
	}
	defer pool.Registry.MustRegister().Unregister()

	graceful.ListenAndServe(httpServer, grpcServer)
}
