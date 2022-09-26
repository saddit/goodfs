package main

import (
	"common/cmd"
	_ "metaserver/cmd/app"
	_ "metaserver/cmd/raft"
	_ "metaserver/cmd/hashslot"
	"os"
)

func main() {
	cmd.Run(os.Args, "app")
}
