package app

import (
	. "apiserver/config"
	"apiserver/internal/controller/http"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/repo"
	"apiserver/internal/usecase/service"
	"common/registry"
	"fmt"
	"os"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

func getLocalAddress() string {
	hn, e := os.Hostname()
	if e != nil {
		panic(e)
	}
	return hn
}

func Run(cfg *Config) {
	pool.InitPool(cfg)
	defer pool.Close()
	// init log
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&nested.Formatter{
		HideKeys:    true,
		FieldsOrder: []string{"component", "category"},
	})
	//init services
	netAddr := fmt.Sprint(getLocalAddress(), ":", cfg.Port)
	versionRepo := repo.NewVersionRepo(pool.Etcd)
	metaRepo := repo.NewMetadataRepo(pool.Etcd, versionRepo)
	metaService := service.NewMetaService(metaRepo, versionRepo)
	objService := service.NewObjectService(metaService, pool.Etcd)
	reg := registry.NewEtcdRegistry(pool.Etcd, cfg.Registry, netAddr)
	// dicovery := registry.NewEtcdDiscovery(pool.Etcd, cfg.Registry.Group, []string{"metaserver", "objectserver"})
	// register self
	err := reg.Register()
	if err != nil {
		logrus.Error(err)
		return
	}
	defer reg.Unregister()
	//start api server
	apiServer := http.NewHttpServer(objService, metaService)
	apiServer.ListenAndServe(netAddr)
}
