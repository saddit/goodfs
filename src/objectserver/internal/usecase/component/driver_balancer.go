package component

import (
	"common/datasize"
	"common/util/slices"
	"errors"
	"math"
	"sort"
)

type DriverBalancer interface {
	Select([]*Driver) (*Driver, error)
}

type spaceFirstBalancer struct{}

func SpaceFirstBalancer() DriverBalancer {
	return &spaceFirstBalancer{}
}

func (sf *spaceFirstBalancer) Select(drivers []*Driver) (*Driver, error) {
	if len(drivers) == 0 {
		return nil, errors.New("non drivers available")
	}
	sort.Slice(drivers, func(i, j int) bool {
		return drivers[i].FreeSpace < drivers[j].FreeSpace
	})
	return slices.Last(drivers), nil
}

type spaceWeightedBalancer struct {
	weighted map[string]datasize.DataSize
	factor   datasize.DataSize
}

func SpaceWeightedBalancer() DriverBalancer {
	return &spaceWeightedBalancer{
		weighted: map[string]datasize.DataSize{},
	}
}

func (sw *spaceWeightedBalancer) initFactor(drivers []*Driver) {
	var total float64
	for _, d := range drivers {
		total += float64(d.TotalSpace)
	}
	avg := math.Log(total / float64(len(drivers)))
	sw.factor = datasize.DataSize(math.Ceil(avg))
}

func (sw *spaceWeightedBalancer) Select(drivers []*Driver) (*Driver, error) {
	if len(drivers) == 0 {
		return nil, errors.New("non drivers available")
	}
	if sw.factor == 0 {
		sw.initFactor(drivers)
	}
	last := slices.Extremal(drivers, func(max, cur *Driver) bool {
		w1 := sw.weighted[max.MountPoint]
		w2 := sw.weighted[cur.MountPoint]
		return cur.FreeSpace-w2 > max.FreeSpace-w1
	})
	if sw.weighted[last.MountPoint] += sw.factor; sw.weighted[last.MountPoint] > last.FreeSpace {
		sw.weighted[last.MountPoint] /= sw.factor
	}
	return last, nil
}
