package raftimpl

import (
	"common/cst"
	"common/graceful"
	"common/logs"
	"common/util"
	"fmt"
	"metaserver/config"
	. "metaserver/internal/usecase"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb/v2"
)

var raftLog = logs.New("raft-impl")

type RaftWrapper struct {
	Raft                *raft.Raft
	ID                  string
	Address             string
	Enabled             bool
	isLeader            bool
	leaderChangedEvents []IRaftLeaderChanged
}

func NewDisabledRaft() *RaftWrapper {
	return &RaftWrapper{Enabled: false}
}

func NewRaft(cfg config.ClusterConfig, fsm raft.FSM, ts raft.Transport) *RaftWrapper {
	baseDir := cfg.StoreDir

	c := raft.DefaultConfig()
	c.LocalID, c.LogOutput, c.LogLevel, c.ElectionTimeout, c.HeartbeatTimeout =
		raft.ServerID(cfg.ID), os.Stderr, cfg.LogLevel, cfg.ElectionTimeout, cfg.HeartbeatTimeout

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
			idAndAddr := strings.Split(v, ",")
			if len(idAndAddr) != 2 {
				panic("raft-bootstrap: nodes item format 'id,host:port'")
			}
			raftCfg.Servers[i] = raft.Server{
				Suffrage: raft.Voter,
				ID:       raft.ServerID(idAndAddr[0]),
				Address:  raft.ServerAddress(idAndAddr[1]),
			}
		}

		go func() {
			f := r.BootstrapCluster(raftCfg)
			if err := f.Error(); err != nil {
				logs.Std().Warnf("raft.Raft.BootstrapCluster: %v", err)
			}
		}()
	}

	rw := &RaftWrapper{Raft: r, ID: cfg.ID, Address: util.GetHostPort(cfg.Port), Enabled: true}
	rw.subscribeLeaderCh()

	return rw
}

//newRaftStore init storage
func newRaftStore(baseDir string) (raft.LogStore, raft.StableStore, raft.SnapshotStore) {
	if err := os.MkdirAll(baseDir, cst.OS.ModeUser); err != nil {
		panic(err)
	}

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

func (rw *RaftWrapper) subscribeLeaderCh() {
	go func() {
		defer graceful.Recover()
		for is := range rw.Raft.LeaderCh() {
			rw.isLeader = is
			rw.OnLeaderChanged(is)
			raftLog.Infof("server %s leader", util.IfElse(is, "become", "lose"))
		}
	}()
}

func (rw *RaftWrapper) OnLeaderChanged(isLeader bool) {
	for _, event := range rw.leaderChangedEvents {
		go func(e IRaftLeaderChanged) {
			defer graceful.Recover()
			e.OnLeaderChanged(isLeader)
		}(event)
	}
}

func (rw *RaftWrapper) Init() {
	// if enabled raft, init as slave
	// else init as singleton master
	rw.OnLeaderChanged(!rw.Enabled)
}

func (rw *RaftWrapper) RegisterLeaderChangedEvent(event IRaftLeaderChanged) {
	rw.leaderChangedEvents = append(rw.leaderChangedEvents, event)
}

func (rw *RaftWrapper) GetRaftIfLeader() (IRaft, bool) {
	if rw.IsLeader() {
		return rw.Raft, true
	}
	return nil, false
}

func (rw *RaftWrapper) IsLeader() bool {
	return rw.isLeader && rw.Enabled
}

func (rw *RaftWrapper) LeaderID() string {
	if !rw.Enabled {
		return ""
	}
	_, id := rw.Raft.LeaderWithID()
	return string(id)
}

func (rw *RaftWrapper) LeaderAddress() string {
	if !rw.Enabled {
		return ""
	}
	addr, _ := rw.Raft.LeaderWithID()
	return string(addr)
}

func (rw *RaftWrapper) Close() error {
	if !rw.Enabled {
		return nil
	}
	raftLog.Info("shutdown raft..")
	return rw.Raft.Shutdown().Error()
}
