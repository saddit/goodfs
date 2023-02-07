package selector

import (
	"apiserver/internal/usecase/webapi"
	"common/logs"
	"common/util"
	"common/util/slices"
	"math"
	"sync"
	"time"
)

var (
	//  Weighted-io is incremented at each I/O start, I/O completion, I/O
	//  merge, or read of these stats by the number of I/Os in progress
	//  times the number of milliseconds spent doing I/O since the
	//  last update of this field.  This can provide an easy measure of both
	//  I/O completion time and the backlog that may be accumulating.
	//  See more: http://www.mjmwired.net/kernel/Documentation/iostats.txt
	weightedIOMap       = map[string]int32{}
	weightedIOLock      = sync.Mutex{}
	weightedIOUpdatedAt time.Time
)

type WeightedIOFirst struct {
}

const IOFirst SelectStrategy = "io-first" // Linux only

func (s *WeightedIOFirst) fetchIoInfo(ds []string) {
	if !weightedIOLock.TryLock() {
		return
	}
	defer weightedIOLock.Unlock()
	for _, ip := range ds {
		hd, err := webapi.StatObject(ip)
		if err != nil {
			logs.Std().Errorf("update space info of %s err: %s", ip, err)
			delete(weightedIOMap, ip)
			continue
		}
		weightedIOMap[ip] = util.ToInt32(hd.Get("Weighted-IO"))
	}
	weightedIOUpdatedAt = time.Now()
}

func (s *WeightedIOFirst) Pop(ds []string) ([]string, string) {
	if time.Since(weightedIOUpdatedAt) > 1*time.Minute {
		s.fetchIoInfo(ds)
	}
	idx := slices.ExtremalIndex(ds, func(target, b string) bool {
		tWeighted, ok := weightedIOMap[target]
		if !ok {
			tWeighted = math.MaxInt32
		}
		bWeighted, ok := weightedIOMap[b]
		if !ok {
			// skip if unreachable
			return false
		}
		return bWeighted < tWeighted
	})
	res := ds[idx]
	ds[0], ds[idx] = ds[idx], ds[0]
	return ds[1:], res
}

func (s *WeightedIOFirst) Select(ds []string) string {
	_, res := s.Pop(ds)
	return res
}

func (s *WeightedIOFirst) Strategy() SelectStrategy {
	return IOFirst
}
