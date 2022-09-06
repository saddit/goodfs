package grpc

import (
	"common/logs"
	"common/util"
	"context"
	"errors"
	"metaserver/config"
	"metaserver/internal/usecase"
	"metaserver/internal/usecase/pb"
	"metaserver/internal/usecase/raftimpl"
	"net"

	raftGrpcService "github.com/Jille/raft-grpc-transport"
	"github.com/hashicorp/raft"
	netGrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var log = logs.New("grpc-server")

type RpcRaftServer struct {
	*netGrpc.Server
	Port string
}

// NewRpcRaftServer init a grpc raft server. if no available nodes return empty object
func NewRpcRaftServer(cfg config.ClusterConfig, repo usecase.IMetadataRepo) (*RpcRaftServer, *raftimpl.RaftWrapper) {
	if len(cfg.Nodes) == 0 {
		log.Warn("no available nodes, raft disabled")
		return &RpcRaftServer{nil, cfg.Port}, raftimpl.NewDisabledRaft()
	}
	server := netGrpc.NewServer(netGrpc.ChainUnaryInterceptor(
		CheckRaftEnabledMid, CheckRaftLeaderMid, CheckRaftNonLeaderMid,
	))
	// init services
	raftGrpcServ := raftGrpcService.New(raft.ServerAddress(util.GetHostPort(cfg.Port)), []netGrpc.DialOption{netGrpc.WithInsecure()})
	raftWrapper := raftimpl.NewRaft(cfg, raftimpl.NewFSM(repo), raftGrpcServ.Transport())
	// register grpc services 
	{
		raftGrpcServ.Register(server)
		reflection.Register(server)
		pb.RegisterRaftCmdServer(server, NewRaftCmdServer(raftWrapper))
	}
	return &RpcRaftServer{server, cfg.Port}, raftWrapper
}

func (r *RpcRaftServer) Shutdown(ctx context.Context) error {
	if r.Server == nil {
		return nil
	}
	finish := make(chan struct{})
	go func () {
		defer close(finish)
		r.Server.GracefulStop()
	}()
	select {
	case <-ctx.Done():
		r.Server.Stop()
		return errors.New("graceful stop grpc server timeout")
	case <-finish:
		return nil
	}
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
