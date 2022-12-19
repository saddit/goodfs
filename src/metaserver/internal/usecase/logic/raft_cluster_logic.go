package logic

import (
	"common/graceful"
	"common/util"
	"metaserver/config"
	"metaserver/internal/usecase/pool"
)

type RaftCluster struct {
}

func NewRaftCluster() *RaftCluster {
	return &RaftCluster{}
}

func (rc RaftCluster) UpdateConfiguration(cfg *config.ClusterConfig) error {
	// update config
	pool.Config.Cluster.GroupID = cfg.GroupID
	pool.Config.Cluster.Nodes = cfg.Nodes
	// re-register peers on change raft cluster
	if err := NewPeers().Register(); err != nil {
		return err
	}

	hsLogic := NewHashSlot()
	if pool.RaftWrapper.IsLeader() {
		// update hash slot group id if is leader
		info, err := hsLogic.GetById(pool.Config.HashSlot.StoreID)
		if err != nil {
			return err
		}
		if err := hsLogic.RemoveFromEtcd(pool.Config.HashSlot.StoreID); err != nil {
			return err
		}
		pool.Config.HashSlot.StoreID = cfg.GroupID
		if err := hsLogic.SaveToEtcd(pool.Config.HashSlot.StoreID, info); err != nil {
			return err
		}
	} else {
		// update config
		pool.Config.HashSlot.StoreID = cfg.GroupID
		info, err := hsLogic.GetById(pool.Config.HashSlot.StoreID)
		if err != nil {
			return err
		}
		pool.Config.HashSlot.Slots = info.Slots
	}

	// persist config async
	go func() {
		defer graceful.Recover()
		util.LogErrWithPre("persist config", pool.Config.Persist())
	}()

	return nil
}
