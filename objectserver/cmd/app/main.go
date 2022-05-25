package main

import (
	"objectserver/config"
	"objectserver/internal/app"
)

func main() {
	conf := config.ReadConfig()
	app.Run(&conf)
}
