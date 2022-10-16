package main

import (
	"common/cmd"
	_ "objectserver/cmd/app"
	_ "objectserver/cmd/rpc"
	"os"
)

func main() {
	cmd.Run(os.Args, "app")
}
