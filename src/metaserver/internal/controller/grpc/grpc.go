package grpc

import (
	"common/logs"
	"common/proto/pb"
	"common/util"
	"context"
	"errors"
	"google.golang.org/grpc"
	"metaserver/internal/usecase"
	"metaserver/internal/usecase/raftimpl"
)

var log = logs.New("grpc-server")

type Server struct {
	*grpc.Server
	leaveCluster func(c context.Context) error
}

// NewRpcServer init a grpc raft server. if no available nodes return empty object
func NewRpcServer(maxStreams uint32, rw *raftimpl.RaftWrapper, serv1 usecase.IMetadataService, serv2 usecase.IHashSlotService, serv3 usecase.BucketService) *Server {
	server := grpc.NewServer(
		util.CommonUnaryInterceptors(),
		util.CommonStreamInterceptors(),
		grpc.MaxConcurrentStreams(maxStreams),
		grpc.ChainUnaryInterceptor(
			CheckKeySlot,
			CheckWritableUnary,
			CheckRaftEnabledUnary,
			CheckRaftLeaderUnary,
			CheckRaftNonLeaderUnary,
		),
		grpc.ChainStreamInterceptor(
			CheckWritableStreaming,
		),
	)
	// init raft service
	leaveRaft := func(context.Context) error { return nil }
	if rw.Enabled {
		rw.Manager.Register(server)
		cmdServer := NewRaftCmdServer(rw)
		leaveRaft = func(c context.Context) error {
			resp, err := cmdServer.LeaveCluster(c, new(pb.EmptyReq))
			if err != nil {
				return err
			}
			if !resp.Success {
				return errors.New(resp.Message)
			}
			return nil
		}
		pb.RegisterRaftCmdServer(server, cmdServer)
	}
	// register services
	// grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	pb.RegisterHashSlotServer(server, NewHashSlotServer(serv2))
	pb.RegisterMetadataApiServer(server, NewMetadataApiServer(serv1, serv3))
	pb.RegisterConfigServiceServer(server, &ConfigServiceServer{})
	return &Server{server, leaveRaft}
}

func (r *Server) Shutdown(ctx context.Context) error {
	if r.Server == nil {
		return nil
	}
	finish := make(chan struct{})
	go func() {
		defer close(finish)
		if err := r.leaveCluster(ctx); err != nil {
			log.Errorf("leave raft-cluster err: %s", err)
		}
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
