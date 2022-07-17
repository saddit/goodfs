package app

import (
	"common/registry"
	. "metaserver/config"
	"metaserver/internal/controller/http"
	"metaserver/internal/usecase/repo"
	"metaserver/internal/usecase/service"
	"os"
	"time"

	"metaserver/internal/controller/grpc"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func Run(cfg *Config) {
	// init logger
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&nested.Formatter{
		HideKeys:    true,
		FieldsOrder: []string{"component", "category"},
	})
	// open db file
	boltdb, err := bolt.Open(cfg.DataDir, os.ModePerm, &bolt.Options{
		Timeout:    12 * time.Second,
		NoGrowSync: false,
	})
	if err != nil {
		logrus.Errorf("open db err: %v", err)
		return
	}
	// init components
	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints: cfg.Etcd.Endpoint,
		Username:  cfg.Etcd.Username,
		Password:  cfg.Etcd.Password,
	})
	if err != nil {
		logrus.Errorf("create etcd client err: %v", err)
		return
	}
	metaRepo := repo.NewMetadataRepo(boltdb)
	metaService := service.NewMetadataService(metaRepo)
	grpcServer := grpc.NewRpcRaftServer(cfg.Cluster, metaService)
	httpServer := http.NewHttpServer(cfg, grpcServer.Server, metaService, grpcServer.Raft)
	reg := registry.NewEtcdRegistry(etcdCli, cfg.Registry, cfg.Cluster.LocalAddr())
	// register self
	if err := reg.Register(); err != nil {
		logrus.Errorf("register err: %v", err)
		return
	}
	defer reg.Unregister()

	httpServer.ListenAndServe()
}
