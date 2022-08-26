package cmd

import (
	"fmt"
	"strings"
	"sync"
)

var (
	cmdStr = ""
	cmdMap = make(map[string]func([]string))
	once   = sync.Once{}
)

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

func Run(args []string, def string) {
	parse()
	if len(args) > 1 {
		def = args[1]
	}
	if len(args) > 2 {
		args = args[2:]
	} else {
		args = args[:0]
	}
	if fn, ok := cmdMap[def]; ok {
		fn(args)
		return
	}
	fmt.Printf("Not found your command '%s'\n", def)
	fmt.Println(cmdStr)
}
