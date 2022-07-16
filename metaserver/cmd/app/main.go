package main

import (
	"metaserver/config"
	"metaserver/internal/app"
)

func main() {
	cfg := config.ReadConfig()
	app.Run(&cfg)
}
