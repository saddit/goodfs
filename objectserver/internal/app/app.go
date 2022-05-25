package app

import (
	"common/graceful"
	"fmt"
	"objectserver/config"

	"objectserver/internal/controller/http"
	"objectserver/internal/controller/amqp"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/service"
	"os"

	"github.com/gin-gonic/gin"
)

func initDir(cfg *config.Config) {
	if !service.ExistPath(cfg.TempPath) {
		if e := os.Mkdir(cfg.TempPath, os.ModeDir); e != nil {
			panic(e)
		}
	}
	if !service.ExistPath(cfg.StoragePath) {
		if e := os.Mkdir(cfg.StoragePath, os.ModeDir); e != nil {
			panic(e)
		}
	}
}

func Run(cfg *config.Config) {
	pool.InitPool(cfg)
	defer pool.Close()

	initDir(cfg)

	router := gin.Default()

	//init router
	http.RegisterRouter(router)
	amqp.Start()
	service.WarmUpLocateCache()
	go service.HandleTempRemove()

	graceful.ListenAndServe(fmt.Sprint(":", cfg.Port), router)
}
