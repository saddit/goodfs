package main

import (
	"common/cmd"
	"os"
	_ "objectserver/cmd/app"
)

func main() {
	cmd.Run(os.Args, "app")
}
