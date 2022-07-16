package grpc

import (
	"metaserver/config"
	. "metaserver/internal/usecase"
	"metaserver/internal/usecase/raftimpl"

	transport "github.com/Jille/raft-grpc-transport"
	"github.com/hashicorp/raft"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type RpcRaftServer struct {
	*ggrpc.Server
	Raft *raft.Raft
}

func NewRpcRaftServer(cfg config.ClusterConfig, service IMetadataService) *RpcRaftServer {
	fsm := raftimpl.NewFSM(service)
	tm := transport.New(raft.ServerAddress(cfg.LocalAddr()), []ggrpc.DialOption{ggrpc.WithInsecure()})
	rf := raftimpl.NewRaft(cfg, fsm, tm.Transport())
	server := ggrpc.NewServer()
	tm.Register(server)
	reflection.Register(server)
	return &RpcRaftServer{server, rf}
}
