package logic

import (
	"common/hashslot"
	"time"
)

type slotCache struct {
	provider  hashslot.IEdgeProvider
	slotIdMap map[string]string
	updatedAt int64
}

func (s *slotCache) update(p hashslot.IEdgeProvider, m map[string]string) {
	s.provider = p
	s.slotIdMap = m
	s.updatedAt = time.Now().Unix()
}

func (s *slotCache) reset() {
	*s = slotCache{}
}

type ipCache struct {
	ips       []string
	updatedAt int64
}

var (
	groupIPCache    = map[string]*ipCache{}
	hashSlotCache   = new(slotCache)
	expiredDuration = int64(time.Minute.Seconds())
)
