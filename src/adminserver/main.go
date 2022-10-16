package main

import (
	_ "adminserver/cmd/app"
	"common/cmd"
	"os"
)

func main() {
	cmd.Run(os.Args, "app")
}
