package raftimpl

import (
	"common/util"
	"fmt"
	"metaserver/config"
	"os"
	"path/filepath"

	raft "github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
)

func NewRaft(cfg config.ClusterConfig, fsm raft.FSM, ts raft.Transport) *raft.Raft {
	addr := util.GetHost()
	baseDir := cfg.StoreDir

	c := raft.DefaultConfig()
	c.LocalID, c.LogOutput, c.LogLevel, c.ElectionTimeout =
		raft.ServerID(addr), os.Stderr, cfg.LogLevel, cfg.ElectionTimeout

	ldb, sdb, fss := newRaftStore(baseDir)

	r, err := raft.NewRaft(c, fsm, ldb, sdb, fss, ts)
	if err != nil {
		panic(fmt.Errorf("raft.NewRaft: %v", err))
	}

	//boot raft cluster
	if cfg.Bootstrap {
		//add existed voter
		raftCfg := raft.Configuration{
			Servers: make([]raft.Server, len(cfg.Nodes)),
		}
		for i, v := range cfg.Nodes {
			raftCfg.Servers[i] = raft.Server{
				Suffrage: raft.Voter,
				ID:       raft.ServerID(v),
				Address:  raft.ServerAddress(v),
			}
		}

		f := r.BootstrapCluster(raftCfg)
		if err := f.Error(); err != nil {
			panic(fmt.Errorf("raft.Raft.BootstrapCluster: %v", err))
		}
	}

	return r
}

//newRaftStore init storage
func newRaftStore(baseDir string) (raft.LogStore, raft.StableStore, raft.SnapshotStore) {

	ldb, err := boltdb.NewBoltStore(filepath.Join(baseDir, "logs.dat"))
	if err != nil {
		panic(fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(baseDir, "logs.dat"), err))
	}

	sdb, err := boltdb.NewBoltStore(filepath.Join(baseDir, "stable.dat"))
	if err != nil {
		panic(fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(baseDir, "stable.dat"), err))
	}

	fss, err := raft.NewFileSnapshotStore(baseDir, 3, os.Stderr)
	if err != nil {
		panic(fmt.Errorf(`raft.NewFileSnapshotStore(%q, ...): %v`, baseDir, err))
	}

	return ldb, sdb, fss
}
