package app

import "common/cmd"
import "apiserver/config"
import "apiserver/internal/app"

func init() {
	cmd.Register("app", func(args []string) {
		conf := config.ReadConfig()
		app.Run(&conf)
	})
}