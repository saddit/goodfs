package db

import (
	"sync/atomic"
)

type ObjectCapacity struct {
	currentCap *atomic.Int64
}

func NewObjectCapacity() *ObjectCapacity {
	return &ObjectCapacity{&atomic.Int64{}}
}

func (oc *ObjectCapacity) AddCap(i int64) {
	oc.currentCap.Add(i)
}

func (oc *ObjectCapacity) SubCap(i int64) {
	oc.currentCap.Add(-i)
}

func (oc *ObjectCapacity) Capacity() int64 {
	return oc.currentCap.Load()
}
