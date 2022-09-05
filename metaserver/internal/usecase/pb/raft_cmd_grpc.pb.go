// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.6.1
// source: raft_cmd.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// RaftCmdClient is the client API for RaftCmd service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RaftCmdClient interface {
	Bootstrap(ctx context.Context, in *BootstrapReq, opts ...grpc.CallOption) (*Response, error)
}

type raftCmdClient struct {
	cc grpc.ClientConnInterface
}

func NewRaftCmdClient(cc grpc.ClientConnInterface) RaftCmdClient {
	return &raftCmdClient{cc}
}

func (c *raftCmdClient) Bootstrap(ctx context.Context, in *BootstrapReq, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/proto.RaftCmd/Bootstrap", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RaftCmdServer is the server API for RaftCmd service.
// All implementations must embed UnimplementedRaftCmdServer
// for forward compatibility
type RaftCmdServer interface {
	Bootstrap(context.Context, *BootstrapReq) (*Response, error)
	mustEmbedUnimplementedRaftCmdServer()
}

// UnimplementedRaftCmdServer must be embedded to have forward compatible implementations.
type UnimplementedRaftCmdServer struct {
}

func (UnimplementedRaftCmdServer) Bootstrap(context.Context, *BootstrapReq) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Bootstrap not implemented")
}
func (UnimplementedRaftCmdServer) mustEmbedUnimplementedRaftCmdServer() {}

// UnsafeRaftCmdServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RaftCmdServer will
// result in compilation errors.
type UnsafeRaftCmdServer interface {
	mustEmbedUnimplementedRaftCmdServer()
}

func RegisterRaftCmdServer(s grpc.ServiceRegistrar, srv RaftCmdServer) {
	s.RegisterService(&RaftCmd_ServiceDesc, srv)
}

func _RaftCmd_Bootstrap_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BootstrapReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftCmdServer).Bootstrap(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RaftCmd/Bootstrap",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftCmdServer).Bootstrap(ctx, req.(*BootstrapReq))
	}
	return interceptor(ctx, in, info, handler)
}

// RaftCmd_ServiceDesc is the grpc.ServiceDesc for RaftCmd service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RaftCmd_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.RaftCmd",
	HandlerType: (*RaftCmdServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Bootstrap",
			Handler:    _RaftCmd_Bootstrap_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "raft_cmd.proto",
}
