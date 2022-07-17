package selector

import (
	"apiserver/internal/entity"
	"log"
	"strings"
)
//TODO 改造成纯IP字符串数组筛选服务器

type SelectStrategy string

type Selector interface {
	Select([]*entity.DataServ) string
	Pop([]*entity.DataServ) ([]*entity.DataServ, string)
	Strategy() SelectStrategy
}

func NewSelector(str string) Selector {
	var sec Selector

	switch strings.ToLower(str) {
	case string(Random):
		sec = &RandomSelector{}
	case string(MaxFreeDisk):
		sec = &MaxFreeDiskSelector{}
	default:
		log.Panicf("Not allowed selector strategy: %v", str)
	}
	return sec
}
