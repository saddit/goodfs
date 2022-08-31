package main

import (
	"common/cmd"
	_ "metaserver/cmd/app"
	"os"
)

func main() {
	cmd.Run(os.Args, "app")
}
