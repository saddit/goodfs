package app

import (
	"adminserver/config"
	"adminserver/internal/controller"
	"adminserver/internal/usecase/pool"
	"adminserver/resource"
	"common/graceful"
)

func Run(cfg *config.Config) {
	pool.Init(cfg)
	defer pool.Close()

	graceful.ListenAndServe(
		nil,
		controller.NewHttpServer(cfg.Port, resource.FileSystem()),
	)
}
