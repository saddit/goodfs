package main

import (
	"metaserver/config"
	"metaserver/internal/app"
	"os"
	"path/filepath"
)

func main() {
	var cfg config.Config
	if len(os.Args) > 1 {
		cfg = config.ReadConfigFrom(os.Args[1])
	} else {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		cfg = config.ReadConfigFrom(filepath.Join(wd, "conf/meta-server.yaml"))
	}
	app.Run(&cfg)
}
