package app

import (
	"common/graceful"
	"metaserver/config"
	"metaserver/internal/controller/grpc"
	"metaserver/internal/controller/http"
	"metaserver/internal/usecase/logic"
	"metaserver/internal/usecase/pool"
	"metaserver/internal/usecase/raftimpl"
	"metaserver/internal/usecase/repo"
	"metaserver/internal/usecase/service"
)

func Run(cfg *config.Config) {
	// init components
	pool.InitPool(cfg)
	defer pool.Close()
	// init services
	metaRepo := repo.NewMetadataRepo(pool.Storage, repo.NewMetadataCacheRepo(pool.Cache))
	bucketRepo := repo.NewBucketRepo(pool.Storage, repo.NewBucketCacheRepo(pool.Cache))
	raftWrapper := raftimpl.NewRaft(cfg.Cluster, raftimpl.NewFSM(metaRepo, bucketRepo))
	bucketServ := service.NewBucketService(bucketRepo, raftWrapper)
	metaService := service.NewMetadataService(
		metaRepo,
		repo.NewBatchRepo(pool.Storage),
		repo.NewHashIndexRepo(pool.Storage),
		raftWrapper,
	)
	hsService := service.NewHashSlotService(pool.HashSlot, metaService, bucketServ, &cfg.HashSlot)
	grpcServer := grpc.NewRpcServer(cfg.RpcPort, raftWrapper, metaService, hsService)
	httpServer := http.NewHttpServer(pool.HttpHostPort, metaService, bucketServ)
	// register on leader change
	raftWrapper.RegisterLeaderChangedEvent(hsService)
	raftWrapper.RegisterLeaderChangedEvent(logic.NewRegistry())
	raftWrapper.Init()
	pool.RaftWrapper = raftWrapper
	// register peers
	defer logic.NewPeers().MustRegister().Unregister()
	// unregister service
	defer pool.Registry.Unregister()
	// auto save disk-info
	defer logic.NewSystemStatLogic().StartAutoSave()()
	// remove slots-info
	defer logic.NewHashSlot().RemoveFromEtcd(cfg.HashSlot.StoreID)
	// flush config
	defer cfg.Persist()

	graceful.ListenAndServe(httpServer, grpcServer)
}
