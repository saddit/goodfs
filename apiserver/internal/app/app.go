package app

import (
	. "apiserver/config"
	"apiserver/internal/controller/amqp"
	"apiserver/internal/controller/http"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/repo"
	"apiserver/internal/usecase/service"
	"apiserver/lib/mongodb"
	"common/graceful"
	"fmt"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/dig"
)

func buildContainer(cfg *Config) *dig.Container {

	container := dig.New()

	container.Provide(service.NewMetaService)
	container.Provide(service.NewObjectService)
	container.Provide(repo.NewMetadataRepo)
	container.Provide(repo.NewVersionRepo)
	container.Provide(func() *mongo.Collection {
		db := mongodb.New(cfg.MongoAddress)
		return db.Collection("metadata_v3")
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

	router := gin.Default()

	buildContainer(cfg).Invoke(func(o *service.ObjectService, m *service.MetaService) {
		http.RegisterRouter(router.Group("/api"), o, m)
		amqp.Start()
	})

	graceful.ListenAndServe(fmt.Sprint(":", cfg.Port), router)
}
