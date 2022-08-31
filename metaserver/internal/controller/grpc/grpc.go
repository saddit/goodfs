package grpc

import (
	"common/logs"
	"common/util"
	"context"
	"metaserver/config"
	"metaserver/internal/usecase"
	"metaserver/internal/usecase/raftimpl"
	"net"

	transport "github.com/Jille/raft-grpc-transport"
	"github.com/hashicorp/raft"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var log = logs.New("grpc-server")

type RpcRaftServer struct {
	*ggrpc.Server
	Port string
}

// NewRpcRaftServer init a grpc raft server. if no available nodes return empty object
func NewRpcRaftServer(cfg config.ClusterConfig, repo usecase.IMetadataRepo) (*RpcRaftServer, *raftimpl.RaftWrapper) {
	if len(cfg.Nodes) == 0 {
		log.Warn("no available nodes, raft disabled")
		return &RpcRaftServer{nil, cfg.Port}, raftimpl.NewDisabledRaft()
	}
	fsm := raftimpl.NewFSM(repo)
	tm := transport.New(raft.ServerAddress(util.GetHostPort(cfg.Port)), []ggrpc.DialOption{ggrpc.WithInsecure()})
	rf := raftimpl.NewRaft(cfg, fsm, tm.Transport())
	server := ggrpc.NewServer()
	tm.Register(server)
	reflection.Register(server)
	return &RpcRaftServer{server, cfg.Port}, rf
}

func (r *RpcRaftServer) Shutdown(ctx context.Context) error {
	if r.Server == nil {
		return nil
	}
	r.Server.GracefulStop()
	return nil
}

func (r *RpcRaftServer) ListenAndServe() error {
	if r.Server == nil {
		log.Warn("server is nil, avoid listening on ", r.Port)
		return nil
	}
	sock, err := net.Listen("tcp", util.GetHostPort(r.Port))
	if err != nil {
		panic(err)
	}
	return r.Serve(sock)
}
