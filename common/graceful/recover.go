package graceful

import (
	"common/logs"
	"fmt"
	"runtime"
	"strings"
)

var logger = logs.New("panic-recover")

func Recover() {
	if err := recover(); err != nil {
		logger.Errorf("%s\n\t%s", err, GetStacks())
	}
}

func GetStacks() string {
	var stack []string
	for i := 1; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		stack = append(stack, fmt.Sprintf("%s:%d", file, line))
	}
	joinStr := ",\r\n"
	return strings.Join(stack, joinStr)
}