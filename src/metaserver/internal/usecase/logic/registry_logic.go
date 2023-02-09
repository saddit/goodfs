package logic

import (
	"common/util"
	"metaserver/internal/usecase/pool"
)

type Registry struct{}

func NewRegistry() Registry { return Registry{} }

func (Registry) OnLeaderChanged(isLeader bool) {
	util.LogErr(pool.Registry.Unregister())
	if isLeader {
		util.LogErr(pool.Registry.AsMaster().Register())
	} else {
		util.LogErr(pool.Registry.AsSlave().Register())
	}
	if lease, ok := pool.Registry.LifecycleLease(); ok {
		pool.HashSlot.Lease = lease
	}
}
