package app

import (
	"common/graceful"
	"common/logs"
	. "metaserver/config"
	"metaserver/internal/controller/grpc"
	"metaserver/internal/controller/http"
	"metaserver/internal/usecase/logic"
	"metaserver/internal/usecase/pool"
	"metaserver/internal/usecase/raftimpl"
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
	var grpcServer *grpc.Server
	metaRepo := repo.NewMetadataRepo(pool.Storage, repo.NewMetadataCacheRepo(pool.Cache))
	metaService := service.NewMetadataService(
		metaRepo,
		repo.NewBatchRepo(pool.Storage),
		repo.NewHashIndexRepo(pool.Storage),
	)
	bucketRepo := repo.NewBucketRepo(pool.Storage, repo.NewBucketCacheRepo(pool.Cache))
	bucketServ := service.NewBucketService(bucketRepo)
	hsService := service.NewHashSlotService(pool.HashSlot, metaService, bucketServ, &cfg.HashSlot)
	grpcServer, pool.RaftWrapper = grpc.NewRpcServer(cfg.Cluster, raftimpl.NewFSM(metaRepo, bucketRepo), metaService, hsService)
	httpServer := http.NewHttpServer(pool.HttpHostPort, metaService, bucketServ)
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
