package grpc

import (
	"context"
	"errors"
	"metaserver/internal/usecase/pool"

	"common/collection/set"

	netGrpc "google.golang.org/grpc"
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

func CheckRaftEnabledMid(ctx context.Context, method string, req, reply interface{}, cc *netGrpc.ClientConn, invoker netGrpc.UnaryInvoker, opts ...netGrpc.CallOption) error {
	if checkRaftEnabledMethods.Contains(method) {
		if !pool.RaftWrapper.Enabled {
			return errors.New("raft is not enabled")
		}
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

func CheckRaftLeaderMid(ctx context.Context, method string, req, reply interface{}, cc *netGrpc.ClientConn, invoker netGrpc.UnaryInvoker, opts ...netGrpc.CallOption) error {
	if checkRaftLeaderMethods.Contains(method) {
		if !pool.RaftWrapper.IsLeader() {
			return errors.New("server is not a leader")
		}
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

func CheckRaftNonLeaderMid(ctx context.Context, method string, req, reply interface{}, cc *netGrpc.ClientConn, invoker netGrpc.UnaryInvoker, opts ...netGrpc.CallOption) error {
	if checkRaftNonLeaderMethods.Contains(method) {
		if pool.RaftWrapper.IsLeader() {
			return errors.New("server is a leader")
		}
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}
