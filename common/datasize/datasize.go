package datasize

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var suffixRegex = regexp.MustCompile(`([.\d]+)(KB|MB|GB|TB|PB)`)

type DataSize int64

const (
	Byte DataSize = 1
	KB            = 1024 * Byte
	MB            = 1024 * KB
	GB            = 1024 * MB
	TB            = 1024 * GB
	PB            = 1024 * TB
)

func (d DataSize) Byte() int64 {
	return int64(d)
}

func (d DataSize) KiloByte() float32 {
	return float32(d * 1.0 / KB)
}

func (d DataSize) MegaByte() float32 {
	return float32(d * 1.0 / MB)
}

func (d DataSize) GigaByte() float32 {
	return float32(d * 1.0 / MB)
}

func (d DataSize) TeraByte() float32 {
	return float32(d * 1.0 / TB)
}

func (d DataSize) PetaByte() float32 {
	return float32(d * 1.0 / PB)
}

var unitNameMap = map[string]DataSize{
	"B": Byte, "KB": KB, "MB": MB,
	"GB": GB, "TB": TB, "PB": PB,
}

func Parse(s string) (DataSize, error) {
	s = strings.ToUpper(s)
	res := suffixRegex.FindAllStringSubmatch(s, 1)
	if len(res) == 0 || len(res[0]) < 3 {
		return 0, fmt.Errorf("data size %v format doesn't support", s)
	}
	num, e := strconv.Atoi(res[0][1])
	if e != nil {
		return 0, e
	}
	if unit, ok := unitNameMap[res[0][2]]; ok {
		return DataSize(num) * unit, nil
	}
	return 0, fmt.Errorf("data size doesn't support unit %v", res[0][2])
}

func MustParse(s string) DataSize {
	r, e := Parse(s)
	if e != nil {
		panic(e)
	}
	return r
}

func (d *DataSize) UnmarshalYAML(node *yaml.Node) error {
	res, err := Parse(node.Value)
	if err != nil {
		return err
	}
	*d = res
	return nil
}
