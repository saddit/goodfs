package selector

import (
	"apiserver/internal/usecase/webapi"
	"common/logs"
	"common/util/slices"
	"sync"
	"time"
)

var (
	spaceUsedMap   = map[string]int64{}
	spaceLock      = sync.Mutex{}
	spaceUpdatedAt time.Time
)

type FreeSpaceFirst struct {
}

const SpaceFirst SelectStrategy = "space-first"

func (s *FreeSpaceFirst) fetchSpaceInfo(ds []string) {
	if !spaceLock.TryLock() {
		return
	}
	defer spaceLock.Unlock()
	for _, ip := range ds {
		size, err := webapi.StatObject(ip)
		if err != nil {
			logs.Std().Errorf("update space info of %s err: %s", ip, err)
			continue
		}
		spaceUsedMap[ip] = size
	}
	spaceUpdatedAt = time.Now()
}

func (s *FreeSpaceFirst) Pop(ds []string) ([]string, string) {
	if time.Since(spaceUpdatedAt) > 5*time.Minute {
		s.fetchSpaceInfo(ds)
	}
	idx := slices.ExtremalIndex(ds, func(max, b string) bool {
		return spaceUsedMap[b] < spaceUsedMap[max]
	})
	res := ds[idx]
	ds[0], ds[idx] = ds[idx], ds[0]
	return ds[1:], res
}

func (s *FreeSpaceFirst) Select(ds []string) string {
	if time.Since(spaceUpdatedAt) > 5*time.Minute {
		s.fetchSpaceInfo(ds)
	}
	return slices.Extremal(ds, func(max, b string) bool {
		return spaceUsedMap[b] < spaceUsedMap[max]
	})
}

func (s *FreeSpaceFirst) Strategy() SelectStrategy {
	return SpaceFirst
}
