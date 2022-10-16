package app

import (
	"common/cmd"
	"metaserver/config"
	"metaserver/internal/app"
	"os"
	"path/filepath"
)

func init() {
	cmd.Register("app", func(args []string) {
		var cfg config.Config
		if len(args) > 0 {
			cfg = config.ReadConfigFrom(args[0])
		} else {
			wd, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			cfg = config.ReadConfigFrom(filepath.Join(wd, "conf/meta-server.yaml"))
		}
		app.Run(&cfg)
	})
}
