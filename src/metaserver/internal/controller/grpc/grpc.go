package grpc

import (
	"common/logs"
	"common/pb"
	"common/util"
	"context"
	"errors"
	"metaserver/internal/usecase"
	"metaserver/internal/usecase/raftimpl"
	"net"

	netGrpc "google.golang.org/grpc"
)

var log = logs.New("grpc-server")

type Server struct {
	*netGrpc.Server
	Port string
}

// NewRpcServer init a grpc raft server. if no available nodes return empty object
func NewRpcServer(port string, rw *raftimpl.RaftWrapper, serv1 usecase.IMetadataService, serv2 usecase.IHashSlotService) *Server {
	server := netGrpc.NewServer(netGrpc.ChainUnaryInterceptor(
		// CheckLocalUnary,
		CheckWritableUnary,
		CheckRaftEnabledUnary,
		CheckRaftLeaderUnary,
		CheckRaftNonLeaderUnary,
		// AllowValidMetaServerUnary,
	), netGrpc.ChainStreamInterceptor(
		CheckWritableStreaming,
		// AllowValidMetaServerStreaming,
	))
	// init raft service
	if rw.Enabled {
		rw.Manager.Register(server)
		pb.RegisterRaftCmdServer(server, NewRaftCmdServer(rw))
	}
	// register hash-slot services
	pb.RegisterHashSlotServer(server, NewHashSlotServer(serv2))
	pb.RegisterMetadataApiServer(server, NewMetadataApiServer(serv1))
	return &Server{server, port}
}

func (r *Server) Shutdown(ctx context.Context) error {
	if r.Server == nil {
		return nil
	}
	finish := make(chan struct{})
	go func() {
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

func (r *Server) ListenAndServe() error {
	if r.Server == nil {
		return nil
	}
	sock, err := net.Listen("tcp", util.GetHostPort(r.Port))
	if err != nil {
		panic(err)
	}
	log.Infof("server listening on %s", sock.Addr().String())
	return r.Serve(sock)
}
