package app

import (
	"common/graceful"
	"metaserver/config"
	"metaserver/internal/controller/grpc"
	"metaserver/internal/controller/http"
	"metaserver/internal/usecase/logic"
	"metaserver/internal/usecase/pool"
	"metaserver/internal/usecase/repo"
	"metaserver/internal/usecase/service"
)

func Run(cfg *config.Config) {
	// init components
	pool.InitPool(cfg)
	defer pool.Close()
	// init services
	var grpcServer *grpc.Server
	metaRepo := repo.NewMetadataRepo(pool.Storage, repo.NewMetadataCacheRepo(pool.Cache))
	metaService := service.NewMetadataService(
		metaRepo,
		repo.NewBatchRepo(pool.Storage),
		repo.NewHashIndexRepo(pool.Storage),
	)
	hsService := service.NewHashSlotService(pool.HashSlot, metaService, &cfg.HashSlot)
	grpcServer, pool.RaftWrapper = grpc.NewRpcServer(cfg.Cluster, metaRepo, metaService, hsService)
	httpServer := http.NewHttpServer(pool.HttpHostPort, metaService)
	// register on leader change
	pool.RaftWrapper.RegisterLeaderChangedEvent(hsService)
	pool.RaftWrapper.RegisterLeaderChangedEvent(logic.NewRegistry())
	pool.RaftWrapper.Init()
	// register peers
	defer logic.NewPeers().MustRegister().Unregister()
	// unregister service
	defer pool.Registry.Unregister()
	// auto save disk-info
	defer logic.NewSystemStatLogic().StartAutoSave()()

	graceful.ListenAndServe(httpServer, grpcServer)
}
