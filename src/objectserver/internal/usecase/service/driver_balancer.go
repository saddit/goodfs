package service

import (
	"common/util/slices"
	"errors"
	"sort"
)

type DriverBalancer interface {
	Select([]*Driver) (*Driver, error)
}

type FreeFirstBalancer struct{}

func NewFreeFirstDriver() *FreeFirstBalancer {
	return &FreeFirstBalancer{}
}

func (ff *FreeFirstBalancer) Select(drivers []*Driver) (*Driver, error) {
	if len(drivers) == 0 {
		return nil, errors.New("non drivers available")
	}
	sort.Slice(drivers, func(i, j int) bool {
		return drivers[i].FreeSpace > drivers[j].FreeSpace
	})
	return slices.Last(drivers), nil
}
