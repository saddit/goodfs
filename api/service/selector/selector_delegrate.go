package selector

import (
	"goodfs/api/config"
	"goodfs/api/model/dataserv"
	"log"
	"strings"
)

type SelectStrategy string

type Selector interface {
	Select([]*dataserv.DataServ) string
	Strategy() SelectStrategy
}

type SelectorDelegrate struct {
	selector Selector
}

func NewSelector(str SelectStrategy) *SelectorDelegrate {
	var sec Selector

	switch strings.ToLower(config.SelectStrategy) {
	case string(Random):
		sec = &RandomSelector{}
	case string(MaxFreeDisk):
		sec = &MaxFreeDiskSelector{}
	default:
		log.Panicf("Not allowed selector strategy %v", str)
	}

	del := &SelectorDelegrate{
		selector: sec,
	}
	return del
}

func (s *SelectorDelegrate) Select(ds []*dataserv.DataServ) string {
	return s.selector.Select(ds)
}
