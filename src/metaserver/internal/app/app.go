package app

import (
	"common/cst"
	"common/graceful"
	"common/system"
	"common/util"
	clientv3 "go.etcd.io/etcd/client/v3"
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
	// flush config
	defer cfg.Persist()
	// init components
	pool.InitPool(cfg)
	defer pool.Close()
	// init repos
	metaRepo := repo.NewMetadataRepo(pool.Storage, repo.NewMetadataCacheRepo(pool.Cache))
	bucketRepo := repo.NewBucketRepo(pool.Storage, repo.NewBucketCacheRepo(pool.Cache))
	// init raft
	fsm := raftimpl.NewFSM(metaRepo, repo.NewBatchRepo(pool.Storage), bucketRepo, repo.NewBatchBucketRepo(pool.Storage), metaRepo)
	raftWrapper := raftimpl.NewRaft(util.ServerAddress(cfg.Port), cfg.Cluster, fsm)
	pool.RaftWrapper = raftWrapper
	// init services
	bucketServ := service.NewBucketService(bucketRepo, raftWrapper)
	metaService := service.NewMetadataService(
		metaRepo,
		repo.NewBatchRepo(pool.Storage),
		repo.NewHashIndexRepo(pool.Storage),
		raftWrapper,
	)
	hsService := service.NewHashSlotService(pool.HashSlot, metaService, bucketServ, &cfg.HashSlot)
	// init server
	grpcServer := grpc.NewRpcServer(cfg.MaxConcurrentStreams, raftWrapper, metaService, hsService, bucketServ)
	httpServer := http.NewHttpServer(cfg.Port, grpcServer, metaService, bucketServ)
	// auto sync sys-info
	syncer := system.Syncer(pool.Etcd, cst.EtcdPrefix.FmtSystemInfo(cfg.Registry.Group, cfg.Registry.Name, cfg.Registry.SID()))
	pool.Lifecycle.Subscribe(func(id clientv3.LeaseID) {
		syncer.LeaseID = id
		_ = syncer.Sync()
	})
	defer syncer.StartAutoSave()()
	// registry
	if raftWrapper.Enabled {
		pool.Registry.AsSlave()
	} else {
		pool.Registry.AsMaster()
	}
	pool.Lifecycle.Subscribe(pool.Registry.Register)
	// unregister service
	defer pool.Registry.Unregister()

	// start lifecycle loop
	go pool.Lifecycle.DeadLoop()

	// register on leader change
	raftWrapper.RegisterLeaderChangedEvent(hsService)
	raftWrapper.RegisterLeaderChangedEvent(logic.NewRegistry())

	// remove slots info from etcd if shutdown as a leader
	defer func() {
		if raftWrapper.IsLeader() || !raftWrapper.Enabled {
			util.LogErr(pool.HashSlot.Remove(cfg.HashSlot.StoreID))
		}
	}()

	graceful.ListenAndServe(nil, httpServer)
}
