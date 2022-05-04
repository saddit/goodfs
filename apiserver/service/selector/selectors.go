package selector

import (
	"goodfs/apiserver/model/dataserv"
	"log"
	"strings"
)

type SelectStrategy string

type Selector interface {
	Select([]*dataserv.DataServ) string
	Pop([]*dataserv.DataServ) ([]*dataserv.DataServ, string)
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
