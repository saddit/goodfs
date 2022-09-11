package main

import (
	"common/cmd"
	_ "objectserver/cmd/app"
	"os"
)

func main() {
	cmd.Run(os.Args, "app")
}
