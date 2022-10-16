package app

import (
	"adminserver/config"
	"adminserver/internal/controller"
	"adminserver/resource"
	"common/graceful"
	"common/util"
)

func Run(cfg *config.Config) {
	httpAddr := util.GetHostPort(cfg.Port)

	graceful.ListenAndServe(
		controller.NewHttpServer(httpAddr, resource.FileSystem()),
	)
}
