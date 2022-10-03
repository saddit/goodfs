package main

import (
	"common/cmd"
	_ "metaserver/cmd/app"
	_ "metaserver/cmd/hashslot"
	_ "metaserver/cmd/raft"
	"os"
)

func main() {
	cmd.Run(os.Args, "app")
}
