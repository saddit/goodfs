package app

import (
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
)

func Run(cfg *Config) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&nested.Formatter{
		HideKeys:    true,
		FieldsOrder: []string{"component", "category"},
	})

	boltdb, e := bolt.Open(cfg.DataDir, os.ModePerm, &bolt.Options{
		Timeout:    12 * time.Second,
		NoGrowSync: false,
	})
	if e != nil {
		panic(e)
	}
	
	metaRepo := repo.NewMetadataRepo(boltdb)
	metaService := service.NewMetadataService(metaRepo)
	grpcServer := grpc.NewRpcRaftServer(cfg.Cluster, metaService)
	httpServer := http.NewHttpServer(cfg, grpcServer.Server, metaService, grpcServer.Raft)

	httpServer.ListenAndServe()

	//TODO 向etcd注册自己
}
