package command

import (
	"flag"
)

type Command struct {
	R string //R repair -r
	A string //A address,ip:port -a
}

func ReadCommand() {
	var cmd Command
	flag.StringVar(&cmd.R, "r", "none", "-r [exist_filter|none]")
	flag.StringVar(&cmd.A, "a", "", "-a ip:port")
	flag.Parse()

	var e error
	if e = RepairCommand(cmd); e != nil {
		flag.PrintDefaults()
		panic(e)
	}
}
