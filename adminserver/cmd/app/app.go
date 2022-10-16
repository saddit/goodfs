package app

import (
	"adminserver/config"
	"adminserver/internal/app"
	"common/cmd"
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
			cfg = config.ReadConfigFrom(filepath.Join(wd, "conf/admin-server.yaml"))
		}
		cfg.ResourcePath, _ = filepath.Abs("./resource")

		app.Run(&cfg)
	})
}
