package raftimpl

import (
	"common/cst"
	"common/graceful"
	"common/logs"
	"common/util"
	"fmt"
	transport "github.com/Jille/raft-grpc-transport"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
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
	Manager             *transport.Manager
	ID                  string
	Address             string
	Enabled             bool
	isLeader            bool
	leaderChangedEvents []IRaftLeaderChanged
	closeRaftStore      func()
}

func NewDisabledRaft() *RaftWrapper {
	return &RaftWrapper{Enabled: false}
}

func NewRaft(localAddr string, cfg config.ClusterConfig, fsm raft.FSM) *RaftWrapper {
	if !cfg.Enable {
		return NewDisabledRaft()
	}
	manager := transport.New(raft.ServerAddress(localAddr), []grpc.DialOption{grpc.WithInsecure()})
	baseDir := cfg.StoreDir

	c := raft.DefaultConfig()
	c.LocalID, c.ElectionTimeout, c.HeartbeatTimeout = raft.ServerID(cfg.ID), cfg.ElectionTimeout, cfg.HeartbeatTimeout
	c.Logger = hclog.New(&hclog.LoggerOptions{
		Name:   "raft",
		Color:  hclog.AutoColor,
		Level:  hclog.LevelFromString(c.LogLevel),
		Output: logs.Std().Out,
	})
	ldb, sdb, fss, closeFn := newRaftStore(baseDir)

	r, err := raft.NewRaft(c, fsm, ldb, sdb, fss, manager.Transport())
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

		if len(raftCfg.Servers) == 0 {
			raftCfg.Servers = append(raftCfg.Servers, raft.Server{
				Suffrage: raft.Voter,
				ID:       raft.ServerID(cfg.ID),
				Address:  raft.ServerAddress(localAddr),
			})
		}

		go func() {
			f := r.BootstrapCluster(raftCfg)
			if err := f.Error(); err != nil {
				logs.Std().Warnf("raft.Raft.BootstrapCluster: %v", err)
			}
		}()
	}

	rw := &RaftWrapper{
		Raft:           r,
		Manager:        manager,
		ID:             cfg.ID,
		Address:        localAddr,
		Enabled:        true,
		closeRaftStore: closeFn,
	}
	rw.subscribeLeaderCh()

	return rw
}

// newRaftStore init storage
func newRaftStore(baseDir string) (raft.LogStore, raft.StableStore, raft.SnapshotStore, func()) {
	if err := os.MkdirAll(baseDir, cst.OS.ModeUser); err != nil {
		panic(err)
	}

	rdb, err := boltdb.NewBoltStore(filepath.Join(baseDir, "raft-store.dat"))
	if err != nil {
		panic(fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(baseDir, "logs.dat"), err))
	}

	fss, err := raft.NewFileSnapshotStore(baseDir, 2, logs.Std().Out)
	if err != nil {
		panic(fmt.Errorf(`raft.NewFileSnapshotStore(%q, ...): %v`, baseDir, err))
	}

	return rdb, rdb, fss, func() { util.LogErr(rdb.Close()) }
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
	defer rw.closeRaftStore()
	raftLog.Info("shutdown raft..")
	return rw.Raft.Shutdown().Error()
}
