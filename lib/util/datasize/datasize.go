package datasize

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type DataSize int

const (
	KB DataSize = 1024
	MB          = 1024 * KB
	GB          = 1024 * MB
	TB          = 1024 * GB
	PB          = 1024 * TB
)

func Parse(s string) (DataSize, error) {
	s = strings.ToUpper(s)
	p, _ := regexp.Compile("(\\d+)(KB|MB|GB|TB|PB)")
	res := p.FindAllStringSubmatch(s, 1)
	if len(res) < 0 && len(res[0]) < 3 {
		return 0, fmt.Errorf("no match this %v", s)
	}
	num, e := strconv.Atoi(res[0][1])
	if e != nil {
		return 0, e
	}
	size := DataSize(num)
	var unit DataSize
	switch res[0][2] {
	case "KB":
		unit = KB
		break
	case "MB":
		unit = MB
		break
	case "GB":
		unit = GB
		break
	case "TB":
		unit = TB
		break
	case "PB":
		unit = PB
		break
	default:
		return 0, fmt.Errorf("no support unit %v", res[1])
	}
	return size * unit, nil
}

func MustParse(s string) DataSize {
	r, e := Parse(s)
	if e != nil {
		panic(e)
	}
	return r
}

func (d DataSize) IntValue() int {
	return int(d)
}

func (d DataSize) Int64Value() int64 {
	return int64(d)
}
