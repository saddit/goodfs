package graceful

import (
	"common/logs"
	"fmt"
	"runtime"
	"strings"
)

var logger = logs.New("panic-recover")

func Recover(fn ...func(msg string)) {
	if err := recover(); err != nil {
		if len(fn) > 0 {
			fn[0](fmt.Sprint(err))
		}
		PrintStacks(err)
	}
}

func PrintStacks(msg any) {
	logger.Errorf("%s\n%s", msg, GetStacks())
}

func GetStacks() string {
	return GetLimitStacks(3, 100)
}

func GetLimitStacks(skip, maxSize int) string {
	var stack []string
	for i := skip; i < maxSize+skip; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		stack = append(stack, fmt.Sprintf("\tat %s:%d", file, line))
	}
	joinStr := "\n"
	return strings.Join(stack, joinStr)
}
