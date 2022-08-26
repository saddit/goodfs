package app

import "objectserver/config"
import "objectserver/internal/app"
import "common/cmd"

func init() {
	cmd.Register("app", func(args []string) {
		conf := config.ReadConfig()
		app.Run(&conf)
	})
}
