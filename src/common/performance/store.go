package performance

import (
	"fmt"
	"strings"
	"time"
)

const (
	ActionWrite = "write"
	ActionRead  = "read"
)

const (
	KindOfHTTP = "http"
	KindOfGRPC = "grpc"
	KindOfBolt = "bolt"
	KindOfDisk = "disk"
)

type Perform struct {
	KindOf string
	Action string
	Cost   time.Duration
}

// Store be used to save or load performance info.
// it may not keep all records since the storage costs
type Store interface {
	// Put store the performance info.
	Put(pm []*Perform) error

	// Get returns all of the performance info determines by both conditions.
	Get(kindOf, action string) ([]*Perform, error)

	// Clear should remove records by both conditions.
	// if both conditions are empty, it should clear all data.
	Clear(kindOf, action string) error

	// Average returns the average of cost.
	// if both conditions are not empty, it should returns only one result.
	Average(kindOf, action string) ([]*Perform, error)

	// Sum returns the sum of cost.
	// if both conditions are not empty, it should returns only one result.
	Sum(kindOf, action string) ([]*Perform, error)

	// Size returns numbers of records by both conditions
	// if both conditions are empty, it should return numbers of all.
	Size(kindOf, action string) (int64, error)
}

type noneStore struct{}

func (ns *noneStore) Put(_ []*Perform) error {
	return nil
}

func (ns *noneStore) Get(_, _ string) ([]*Perform, error) {
	return []*Perform{}, nil
}

func (ns *noneStore) Average(_, _ string) ([]*Perform, error) {
	return []*Perform{}, nil
}

func (ns *noneStore) Sum(_, _ string) ([]*Perform, error) {
	return []*Perform{}, nil
}

func (ns *noneStore) Clear(_, _ string) error {
	return nil
}

func (ns *noneStore) Size(_, _ string) (int64, error) {
	return 0, nil
}

func NoneStore() Store {
	return &noneStore{}
}

type avgSumStore struct {
	Store
}

// AvgSumStore inner store could just implement Get, Put, Clear and Size.
// this will get data from inner store and calculate their Average and Sum.
func AvgSumStore(inner Store) Store {
	return &avgSumStore{inner}
}

func (as *avgSumStore) Average(kind, action string) ([]*Perform, error) {
	data, err := as.Get(kind, action)
	if err != nil {
		return nil, err
	}
	totalMap := make(map[string]time.Duration)
	sizeMap := make(map[string]int)
	for _, item := range data {
		key := fmt.Sprint(item.KindOf, ".", item.Action)
		totalMap[key] += item.Cost
		sizeMap[key]++
	}
	res := make([]*Perform, 0, len(totalMap))
	for k, total := range totalMap {
		sp := strings.Split(k, ".")
		kind, action := sp[0], sp[1]
		res = append(res, &Perform{
			KindOf: kind,
			Action: action,
			Cost:   total / time.Duration(sizeMap[k]),
		})
	}
	return res, nil
}

func (as *avgSumStore) Sum(kind, action string) ([]*Perform, error) {
	data, err := as.Get(kind, action)
	if err != nil {
		return nil, err
	}
	totalMap := make(map[string]time.Duration)
	for _, item := range data {
		key := fmt.Sprint(item.KindOf, ".", item.Action)
		totalMap[key] += item.Cost
	}
	res := make([]*Perform, 0, len(totalMap))
	for k, total := range totalMap {
		sp := strings.Split(k, ".")
		kind, action := sp[0], sp[1]
		res = append(res, &Perform{
			KindOf: kind,
			Action: action,
			Cost:   total,
		})
	}
	return res, nil
}
