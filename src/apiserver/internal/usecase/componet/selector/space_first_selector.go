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
		hd, err := webapi.StatObject(ip)
		if err != nil {
			logs.Std().Errorf("update space info of %s err: %s", ip, err)
			// ip is unreachable remove from origin
			delete(spaceUsedMap, ip)
			continue
		}
		spaceUsedMap[ip] = util.ToInt64(hd.Get("Capacity"))
	}
	spaceUpdatedAt = time.Now()
}

func (s *FreeSpaceFirst) Pop(ds []string) ([]string, string) {
	if time.Since(spaceUpdatedAt) > 1*time.Minute {
		s.fetchSpaceInfo(ds)
	}
	idx := slices.ExtremalIndex(ds, func(min, b string) bool {
		minUsed, ok := spaceUsedMap[min]
		if !ok {
			minUsed = math.MaxInt64
		}
		bUsed, ok := spaceUsedMap[b]
		if !ok {
			// skip if b is unreachable
			return false
		}
		return bUsed < minUsed
	})
	res := ds[idx]
	ds[0], ds[idx] = ds[idx], ds[0]
	return ds[1:], res
}

func (s *FreeSpaceFirst) Select(ds []string) string {
	if time.Since(spaceUpdatedAt) > 1*time.Minute {
		s.fetchSpaceInfo(ds)
	}
	return slices.Extremal(ds, func(min, b string) bool {
		minUsed, ok := spaceUsedMap[min]
		if !ok {
			minUsed = math.MaxInt64
		}
		bUsed, ok := spaceUsedMap[b]
		if !ok {
			// skip if b is unreachable
			return false
		}
		return bUsed < minUsed
	})
}

func (s *FreeSpaceFirst) Strategy() SelectStrategy {
	return SpaceFirst
}
