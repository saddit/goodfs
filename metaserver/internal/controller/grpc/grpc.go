package grpc

import (
	"common/util"
	"metaserver/config"
	. "metaserver/internal/usecase"
	"metaserver/internal/usecase/raftimpl"

	transport "github.com/Jille/raft-grpc-transport"
	"github.com/hashicorp/raft"
	"github.com/sirupsen/logrus"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type RpcRaftServer struct {
	*ggrpc.Server
	Raft *raft.Raft
}

// NewRpcRaftServer init a grpc raft server. if no available nodes return empty object
func NewRpcRaftServer(cfg config.ClusterConfig, tx ITransaction) *RpcRaftServer {
	if len(cfg.Nodes) == 0 {
		logrus.Warn("no available nodes, raft disabled")
		return &RpcRaftServer{nil, nil}
	}
	fsm := raftimpl.NewFSM(tx)
	tm := transport.New(raft.ServerAddress(util.GetHost()), []ggrpc.DialOption{ggrpc.WithInsecure()})
	rf := raftimpl.NewRaft(cfg, fsm, tm.Transport())
	server := ggrpc.NewServer()
	tm.Register(server)
	reflection.Register(server)
	return &RpcRaftServer{server, rf}
}
