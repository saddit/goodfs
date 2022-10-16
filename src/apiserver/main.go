package main

import (
	_ "apiserver/cmd/app"
	"common/cmd"
	"os"
)

func main() {
	cmd.Run(os.Args, "app")
}
