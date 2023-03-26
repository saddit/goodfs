package logic

import (
	"metaserver/internal/usecase/pool"
)

type Registry struct{}

func NewRegistry() Registry { return Registry{} }

func (Registry) OnLeaderChanged(isLeader bool) {
	if isLeader {
		pool.Registry.AsMaster().Register(pool.Lifecycle.Lease())
	} else {
		pool.Registry.AsSlave().Register(pool.Lifecycle.Lease())
	}
}
