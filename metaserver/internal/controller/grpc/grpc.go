package grpc

import (
	"common/logs"
	"common/util"
	"context"
	"metaserver/config"
	"metaserver/internal/usecase/db"
	"metaserver/internal/usecase/raftimpl"
	"net"

	transport "github.com/Jille/raft-grpc-transport"
	"github.com/hashicorp/raft"
	"github.com/sirupsen/logrus"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	log = logs.New("rpc-raft")
)

type RpcRaftServer struct {
	*ggrpc.Server
	Raft *raft.Raft
	cfg config.ClusterConfig
}

// NewRpcRaftServer init a grpc raft server. if no available nodes return empty object
func NewRpcRaftServer(cfg config.ClusterConfig, tx *db.Storage) *RpcRaftServer {
	if len(cfg.Nodes) == 0 {
		logrus.Warn("no available nodes, raft disabled")
		return &RpcRaftServer{nil, nil, cfg}
	}
	fsm := raftimpl.NewFSM(tx)
	tm := transport.New(raft.ServerAddress(util.GetHostPort(cfg.Port)), []ggrpc.DialOption{ggrpc.WithInsecure()})
	rf := raftimpl.NewRaft(cfg, fsm, tm.Transport())
	server := ggrpc.NewServer()
	tm.Register(server)
	reflection.Register(server)
	return &RpcRaftServer{server, rf, cfg}
}

func (r *RpcRaftServer) Shutdown(ctx context.Context) error {
	defer r.Server.GracefulStop()
	if err := r.Raft.Shutdown().Error(); err != nil {
		return err
	}
	return nil
}

func (r *RpcRaftServer) ListenAndServe() error {
	sock, err := net.Listen("tcp", util.GetHostPort(r.cfg.Port))
	if err != nil {
		panic(err)
	}
	return r.Serve(sock)
}
