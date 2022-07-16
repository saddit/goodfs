package app

import (
	"apiserver/config"
	. "apiserver/config"
	"apiserver/internal/controller/amqp"
	"apiserver/internal/controller/http"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/repo"
	"apiserver/internal/usecase/service"
	"common/graceful"
	"fmt"
	"time"

	"github.com/838239178/goodmq"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/dig"
)

func buildContainer(cfg *Config) *dig.Container {

	container := dig.New()

	container.Provide(service.NewMetaService)
	container.Provide(service.NewObjectService)
	container.Provide(repo.NewMetadataRepo)
	container.Provide(repo.NewVersionRepo)
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
	//TODO provide Registry and Discovery

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

	router := gin.Default()

	container := buildContainer(cfg)

	container.Invoke(func(
		o *service.ObjectService,
		m *service.MetaService,
		conn *goodmq.AmqpConnection,
	) {
		http.RegisterRouter(router.Group("/api"), o, m)
		amqp.Start(cfg.Discovery, conn)
		//TODO 向etcd注册自己
	})

	graceful.ListenAndServe(fmt.Sprint(":", cfg.Port), router)

	//do release before quit
	container.Invoke(func(
		conn *goodmq.AmqpConnection,
		prov *goodmq.AmqpProvider,
		etcd *clientv3.Client,
	) {
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
}
