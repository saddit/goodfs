package app

import (
	"apiserver/config"
	. "apiserver/config"
	"apiserver/internal/controller/http"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/repo"
	"apiserver/internal/usecase/service"
	"common/registry"
	"fmt"
	"os"
	"time"

	"github.com/838239178/goodmq"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/dig"
)

func getLocalAddress() string {
	hn, e := os.Hostname()
	if e != nil {
		panic(e)
	}
	return hn
}

func buildContainer(cfg *Config) *dig.Container {

	container := dig.New()

	container.Provide(service.NewMetaService)
	container.Provide(service.NewObjectService)
	container.Provide(repo.NewMetadataRepo)
	container.Provide(repo.NewVersionRepo)
	container.Provide(http.NewHttpServer)
	container.Provide(func(cfg *config.Config) *goodmq.AmqpConnection {
		goodmq.RecoverDelay = 3 * time.Second
		return goodmq.NewAmqpConnection(cfg.AmqpAddress)
	})
	container.Provide(func(conn *goodmq.AmqpConnection) *goodmq.AmqpProvider {
		prov, e := conn.NewProvider()
		if e != nil {
			panic(e)
		}
		return prov
	})
	container.Provide(func() *clientv3.Client {
		c, e := clientv3.New(clientv3.Config{
			Endpoints: cfg.Etcd.Endpoint,
			Username:  cfg.Etcd.Username,
			Password:  cfg.Etcd.Password,
		})
		if e != nil {
			panic(e)
		}
		return c
	})
	container.Provide(func(cli *clientv3.Client) registry.Registry {
		return registry.NewEtcdRegistry(cli, cfg.Registry, getLocalAddress())
	})
	container.Provide(func(cli *clientv3.Client) registry.Discovery {
		return registry.NewEtcdDiscovery(cli, cfg.Registry.Group, []string{"metaserver", "objectserver"})
	})

	return container
}

func Run(cfg *Config) {
	pool.InitPool(cfg)
	defer pool.Close()

	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&nested.Formatter{
		HideKeys:    true,
		FieldsOrder: []string{"component", "category"},
	})

	container := buildContainer(cfg)
	//initialize
	err := container.Invoke(func(
		reg registry.Registry,
	) error {
		if err := reg.Register(); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logrus.Error(err)
		return
	}
	//do release before quit
	defer container.Invoke(func(
		conn *goodmq.AmqpConnection,
		prov *goodmq.AmqpProvider,
		etcd *clientv3.Client,
		reg registry.Registry,
	) {
		if e := reg.Unregister(); e != nil {
			logrus.Error(e)
		}
		if e := conn.Close(); e != nil {
			logrus.Error(e)
		}
		if e := prov.Close(); e != nil {
			logrus.Error(e)
		}
		if e := etcd.Close(); e != nil {
			logrus.Error(e)
		}
	})
	//start api server
	container.Invoke(func(server *http.HttpServer) {
		server.ListenAndServe(fmt.Sprint(":", cfg.Port))
	})
}
