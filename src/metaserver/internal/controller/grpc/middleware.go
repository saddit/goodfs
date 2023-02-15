package grpc

import (
	"common/collection/set"
	"common/graceful"
	"common/proto/pb"
	"context"
	"metaserver/internal/usecase/logic"
	"metaserver/internal/usecase/pool"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var checkRaftEnabledMethods = set.OfString([]string{
	"/proto.RaftCmd/Bootstrap",
	"/proto.RaftCmd/AddVoter",
	"/proto.RaftCmd/RemoveFollower",
	"/proto.RaftCmd/LeaveCluster",
})

var checkRaftLeaderMethods = set.OfString([]string{
	"/proto.RaftCmd/AddVoter",
	"/proto.RaftCmd/RemoveFollower",
})

var checkRaftNonLeaderMethods = set.OfString([]string{
	"/proto.RaftCmd/JoinLeader",
})

//var checkLocalMethods = set.OfString([]string{
//	"/proto.RaftCmd/Bootstrap",
//	"/proto.RaftCmd/AddVoter",
//	"/proto.RaftCmd/JoinLeader",
//	"/proto.RaftCmd/AddVoter",
//	"/proto.HashSlot/StartMigration",
//	"/proto.HashSlot/GetCurrentSlots",
//})

var checkWritableMethods = set.OfString([]string{
	"/proto.HashSlot/StartMigration",
	"/proto.HashSlot/PrepareMigration",
	"/proto.HashSlot/StreamingReceive",
})

//var checkValidMetaServerMethods = set.OfString([]string{
//	"/proto.HashSlot/PrepareMigration",
//	"/proto.HashSlot/StreamingReceive",
//})

func CheckRaftEnabledUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if checkRaftEnabledMethods.Contains(info.FullMethod) {
		if !pool.RaftWrapper.Enabled {
			return nil, status.Error(codes.Unavailable, "raft is not enabled")
		}
	}
	return handler(ctx, req)
}

func CheckRaftLeaderUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if checkRaftLeaderMethods.Contains(info.FullMethod) {
		if !pool.RaftWrapper.IsLeader() {
			return nil, status.Error(codes.Unavailable, "server is not a leader")
		}
	}
	return handler(ctx, req)
}

func CheckWritableUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if checkWritableMethods.Contains(info.FullMethod) {
		if pool.RaftWrapper.Enabled && !pool.RaftWrapper.IsLeader() {
			return nil, status.Error(codes.Unavailable, "server is not writable")
		}
	}
	return handler(ctx, req)
}

func CheckWritableStreaming(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if checkWritableMethods.Contains(info.FullMethod) {
		if pool.RaftWrapper.Enabled && !pool.RaftWrapper.IsLeader() {
			return status.Error(codes.Unavailable, "server is not writable")
		}
	}
	return handler(srv, ss)
}

func CheckRaftNonLeaderUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if checkRaftNonLeaderMethods.Contains(info.FullMethod) {
		if pool.RaftWrapper.IsLeader() {
			return nil, status.Error(codes.Unavailable, "server is a leader")
		}
	}
	return handler(ctx, req)
}

func CheckKeySlot(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request value can not be nil")
	}
	r, ok := req.(*pb.MetaReq)
	if !ok {
		if err := checkKeySlotMetadata(req); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
	if r.Id == "" {
		return handler(ctx, req)
	}
	ok, other := logic.NewHashSlot().IsKeyOnThisServer(r.Id)
	if ok {
		return handler(ctx, req)
	}
	return nil, status.Error(codes.Aborted, logic.NewDiscovery().PeerIp(other, true))
}

func checkKeySlotMetadata(req interface{}) error {
	r, ok := req.(*pb.Metadata)
	if !ok {
		return nil
	}
	if r.Id == "" {
		return nil
	}
	ok, other := logic.NewHashSlot().IsKeyOnThisServer(r.Id)
	if ok {
		return nil
	}
	return status.Error(codes.Aborted, logic.NewDiscovery().PeerIp(other, true))
}

// UnaryServerRecoveryInterceptor returns a new unary server interceptor for panic recovery.
func UnaryServerRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		defer graceful.Recover(func(msg string) {
			err = status.Error(codes.Internal, "panic")
		})
		return handler(ctx, req)
	}
}

// StreamServerRecoveryInterceptor returns a new streaming server interceptor for panic recovery.
func StreamServerRecoveryInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer graceful.Recover(func(msg string) {
			err = status.Error(codes.Internal, "panic")
		})
		return handler(srv, stream)
	}
}
