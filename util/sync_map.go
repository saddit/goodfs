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

//Get Directly return the pointer of value
//important: Every change in v will affect to original value
func (m *SyncMap[K, V]) Get(key K) (*V, bool) {
	v, ok := m.mp.Load(key)
	if ok {
		return v.(*V), ok
	} else {
		return nil, ok
	}
}

//Get2 Copy value to v
//important: Every change to v will not affect original value
func (m *SyncMap[K, V]) Get2(key K, v *V) bool {
	if val, ok := m.mp.Load(key); ok {
		if temp, ok := val.(*V); ok {
			v = temp
			return true
		}
	}
	return false
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
