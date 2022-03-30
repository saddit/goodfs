package util

import (
	"sync"
)

type SyncMap[K interface{}, V interface{}] struct {
	mp *sync.Map
}

func NewSyncMap[K interface{}, V interface{}]() *SyncMap[K, V] {
	syncMap := SyncMap[K, V]{
		mp: &sync.Map{},
	}

	return &syncMap
}

func (m *SyncMap[K, V]) Get(key K) (*V, bool) {
	v, ok := m.mp.Load(key)
	if ok {
		return v.(*V), ok
	} else {
		return nil, ok
	}

}

func (m *SyncMap[K, V]) ForEach(f func(key K, value *V)) {
	m.mp.Range(func(key, value any) bool {
		f(key.(K), value.(*V))
		return true
	})
}

func (m *SyncMap[K, V]) Remove(key K) (*V, bool) {
	v, ok := m.mp.LoadAndDelete(key)
	if ok {
		return v.(*V), ok
	} else {
		return nil, ok
	}
}

func (m *SyncMap[K, V]) Put(key K, value *V) {
	m.mp.Store(key, value)
}

func (m *SyncMap[K, V]) Contains(key K) bool {
	_, ok := m.mp.Load(key)
	return ok
}
