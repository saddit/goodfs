package app

import (
	"common/graceful"
	"common/logs"
	"common/registry"
	"common/util"
	. "metaserver/config"
	"metaserver/internal/controller/grpc"
	"metaserver/internal/controller/http"
	"metaserver/internal/usecase/db"
	"metaserver/internal/usecase/repo"
	"metaserver/internal/usecase/service"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var logger = logs.Std()

func Run(cfg *Config) {
	// init logger
	logs.SetLevel(cfg.LogLevel)
	// open db file
	boltdb := db.NewStorage()
	if err := boltdb.Open(cfg.DataDir); err != nil {
		logger.Errorf("open db err: %v", err)
		return
	}
	defer boltdb.Close()
	// init components
	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints: cfg.Etcd.Endpoint,
		Username:  cfg.Etcd.Username,
		Password:  cfg.Etcd.Password,
	})
	if err != nil {
		logger.Errorf("create etcd client err: %v", err)
		return
	}
	netAddr := util.GetHostPort(cfg.Port)
	metaRepo := repo.NewMetadataRepo(boltdb)
	metaService := service.NewMetadataService(metaRepo)
	grpcServer := grpc.NewRpcRaftServer(cfg.Cluster, boltdb)
	metaRepo.Raft = grpcServer.Raft
	httpServer := http.NewHttpServer(netAddr, metaService)
	defer registry.NewEtcdRegistry(etcdCli, cfg.Registry, netAddr).MustRegister().Unregister()
	graceful.ListenAndServe(httpServer, grpcServer)
}
