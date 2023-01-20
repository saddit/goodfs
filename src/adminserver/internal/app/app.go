package app

import (
	"adminserver/config"
	"adminserver/internal/controller"
	"adminserver/internal/usecase/pool"
	"adminserver/resource"
	"common/graceful"
	"common/util"
)

func Run(cfg *config.Config) {
	pool.Init(cfg)
	defer pool.Close()

	httpAddr := util.GetHostPort(cfg.Port)

	graceful.ListenAndServe(
		nil,
		controller.NewHttpServer(httpAddr, resource.FileSystem()),
	)
}
