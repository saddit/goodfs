package main

import (
	"metaserver/config"
	"metaserver/internal/app"
	"os"
	"path/filepath"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cfg := config.ReadConfigFrom(filepath.Join(wd, "conf/meta-server.yaml"))
	app.Run(&cfg)
}
