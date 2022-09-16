package grpc

import (
	"context"
	"errors"
	"metaserver/internal/usecase/pool"
	"common/util"
	"common/collection/set"

	netGrpc "google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

var checkRaftEnabledMethods = set.OfString([]string{
	"/proto.RaftCmd/Bootstrap",
	"/proto.RaftCmd/AddVoter",
})

var checkRaftLeaderMethods = set.OfString([]string {
	"/proto.RaftCmd/AddVoter",
})

var checkRaftNonLeaderMethods = set.OfString([]string {
	"/proto.RaftCmd/JoinLeader",
})

var checkLocalMethods = set.OfString([]string{
	"/proto.RaftCmd/Bootstrap",
	"/proto.RaftCmd/AddVoter",
	"/proto.RaftCmd/JoinLeader",
	"/proto.RaftCmd/AddVoter",
})

func CheckRaftEnabledMid(ctx context.Context, req interface{}, info *netGrpc.UnaryServerInfo, handler netGrpc.UnaryHandler) (any, error) {
	if checkRaftEnabledMethods.Contains(info.FullMethod) {
		if !pool.RaftWrapper.Enabled {
			return nil, errors.New("raft is not enabled")
		}
	}
	return handler(ctx, req)
}

func CheckRaftLeaderMid(ctx context.Context, req interface{}, info *netGrpc.UnaryServerInfo, handler netGrpc.UnaryHandler) (any, error) {
	if checkRaftLeaderMethods.Contains(info.FullMethod) {
		if !pool.RaftWrapper.IsLeader() {
			return nil, errors.New("server is not a leader")
		}
	}
	return handler(ctx, req)
}

func CheckRaftNonLeaderMid(ctx context.Context, req interface{}, info *netGrpc.UnaryServerInfo, handler netGrpc.UnaryHandler) (any, error) {
	if checkRaftNonLeaderMethods.Contains(info.FullMethod) {
		if pool.RaftWrapper.IsLeader() {
			return nil, errors.New("server is a leader")
		}
	}
	return handler(ctx, req)
}

func CheckLocalMid(ctx context.Context, req interface{}, info *netGrpc.UnaryServerInfo, handler netGrpc.UnaryHandler) (any, error) {
	if checkLocalMethods.Contains(info.FullMethod) {
		if pr, ok := peer.FromContext(ctx); ok {
			if util.GetHostFromAddr(pr.Addr.String()) != util.GetHost() {
				return nil, errors.New("deny")
			}
		} else {
			return nil, errors.New("get client ip fail")
		}
	}
	return handler(ctx, req)
}