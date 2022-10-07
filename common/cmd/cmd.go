package cmd

import (
	"fmt"
	"strings"
	"sync"
)

var (
	cmdStr = ""
	cmdMap = make(map[string]CommandFunc)
	once   = sync.Once{}
)

type CommandFunc func([]string)

func parse() {
	once.Do(func() {
		b := strings.Builder{}
		b.WriteString("support commands:\n")
		cnt := 1
		for k := range cmdMap {
			b.WriteString(fmt.Sprintf("  %d. %s\n", cnt, k))
			cnt++
		}
		cmdStr = b.String()
	})
}

func Register(name string, fn func(args []string)) {
	cmdMap[name] = fn
}

func Run(args []string, def ...string) {
	parse()
	if len(args) <= 1 {
		args = append(args, def...)
	}
	// get app cmd
	cmd := args[1]
	// clip args leaving app-cmd args
	if len(args) > 2 {
		args = args[2:]
	} else {
		args = args[:0]
	}
	// run cmd if exist
	if fn, ok := cmdMap[cmd]; ok {
		fn(args)
		return
	}
	// print hint if not found cmd
	fmt.Printf("Not found your command '%s'\n", def)
	fmt.Println(cmdStr)
}
