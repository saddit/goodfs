package set

import "sync"

type MapSet struct {
	mp map[any]struct{}
}

func (ms *MapSet) Add(elem any) {
	ms.mp[elem] = struct{}{}
}

func (ms *MapSet) Remove(elem any) bool {
	if ms.Contains(elem) {
		delete(ms.mp, elem)
		return true
	}
	return false
}

func (ms *MapSet) Contains(elem any) bool {
	_, ok := ms.mp[elem]
	return ok
}

func (ms *MapSet) Size() int {
	return len(ms.mp)
}

func (ms *MapSet) Foreach(fn func(elem any)) {
	for k := range ms.mp {
		fn(k)
	}
}

func NewMapSet() *MapSet {
	return &MapSet{mp: make(map[any]struct{})}
}

type WriteSyncSet struct {
	*MapSet
	mux sync.Mutex
}

func NewWriteSyncSet() *WriteSyncSet {
	return &WriteSyncSet{MapSet: NewMapSet()}
}

func (wss *WriteSyncSet) Add(elem any) {
	wss.mux.Lock()
	defer wss.mux.Unlock()
	wss.MapSet.Add(elem)
}

func (wss *WriteSyncSet) Remove(elem any) bool {
	wss.mux.Lock()
	defer wss.mux.Unlock()
	return wss.MapSet.Remove(elem)
}
