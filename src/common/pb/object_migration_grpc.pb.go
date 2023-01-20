// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.5
// source: object_migration.proto

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

// ObjectMigrationClient is the client API for ObjectMigration service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ObjectMigrationClient interface {
	ReceiveData(ctx context.Context, opts ...grpc.CallOption) (ObjectMigration_ReceiveDataClient, error)
	FinishReceive(ctx context.Context, in *ObjectInfo, opts ...grpc.CallOption) (*Response, error)
	RequireSend(ctx context.Context, in *RequiredInfo, opts ...grpc.CallOption) (*Response, error)
	LeaveCommand(ctx context.Context, in *EmptyReq, opts ...grpc.CallOption) (*Response, error)
	JoinCommand(ctx context.Context, in *EmptyReq, opts ...grpc.CallOption) (*Response, error)
}

type objectMigrationClient struct {
	cc grpc.ClientConnInterface
}

func NewObjectMigrationClient(cc grpc.ClientConnInterface) ObjectMigrationClient {
	return &objectMigrationClient{cc}
}

func (c *objectMigrationClient) ReceiveData(ctx context.Context, opts ...grpc.CallOption) (ObjectMigration_ReceiveDataClient, error) {
	stream, err := c.cc.NewStream(ctx, &ObjectMigration_ServiceDesc.Streams[0], "/proto.ObjectMigration/ReceiveData", opts...)
	if err != nil {
		return nil, err
	}
	x := &objectMigrationReceiveDataClient{stream}
	return x, nil
}

type ObjectMigration_ReceiveDataClient interface {
	Send(*ObjectData) error
	CloseAndRecv() (*Response, error)
	grpc.ClientStream
}

type objectMigrationReceiveDataClient struct {
	grpc.ClientStream
}

func (x *objectMigrationReceiveDataClient) Send(m *ObjectData) error {
	return x.ClientStream.SendMsg(m)
}

func (x *objectMigrationReceiveDataClient) CloseAndRecv() (*Response, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *objectMigrationClient) FinishReceive(ctx context.Context, in *ObjectInfo, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/proto.ObjectMigration/FinishReceive", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectMigrationClient) RequireSend(ctx context.Context, in *RequiredInfo, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/proto.ObjectMigration/RequireSend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectMigrationClient) LeaveCommand(ctx context.Context, in *EmptyReq, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/proto.ObjectMigration/LeaveCommand", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectMigrationClient) JoinCommand(ctx context.Context, in *EmptyReq, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/proto.ObjectMigration/JoinCommand", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ObjectMigrationServer is the server API for ObjectMigration service.
// All implementations must embed UnimplementedObjectMigrationServer
// for forward compatibility
type ObjectMigrationServer interface {
	ReceiveData(ObjectMigration_ReceiveDataServer) error
	FinishReceive(context.Context, *ObjectInfo) (*Response, error)
	RequireSend(context.Context, *RequiredInfo) (*Response, error)
	LeaveCommand(context.Context, *EmptyReq) (*Response, error)
	JoinCommand(context.Context, *EmptyReq) (*Response, error)
	mustEmbedUnimplementedObjectMigrationServer()
}

// UnimplementedObjectMigrationServer must be embedded to have forward compatible implementations.
type UnimplementedObjectMigrationServer struct {
}

func (UnimplementedObjectMigrationServer) ReceiveData(ObjectMigration_ReceiveDataServer) error {
	return status.Errorf(codes.Unimplemented, "method ReceiveData not implemented")
}
func (UnimplementedObjectMigrationServer) FinishReceive(context.Context, *ObjectInfo) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FinishReceive not implemented")
}
func (UnimplementedObjectMigrationServer) RequireSend(context.Context, *RequiredInfo) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequireSend not implemented")
}
func (UnimplementedObjectMigrationServer) LeaveCommand(context.Context, *EmptyReq) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LeaveCommand not implemented")
}
func (UnimplementedObjectMigrationServer) JoinCommand(context.Context, *EmptyReq) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method JoinCommand not implemented")
}
func (UnimplementedObjectMigrationServer) mustEmbedUnimplementedObjectMigrationServer() {}

// UnsafeObjectMigrationServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ObjectMigrationServer will
// result in compilation errors.
type UnsafeObjectMigrationServer interface {
	mustEmbedUnimplementedObjectMigrationServer()
}

func RegisterObjectMigrationServer(s grpc.ServiceRegistrar, srv ObjectMigrationServer) {
	s.RegisterService(&ObjectMigration_ServiceDesc, srv)
}

func _ObjectMigration_ReceiveData_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ObjectMigrationServer).ReceiveData(&objectMigrationReceiveDataServer{stream})
}

type ObjectMigration_ReceiveDataServer interface {
	SendAndClose(*Response) error
	Recv() (*ObjectData, error)
	grpc.ServerStream
}

type objectMigrationReceiveDataServer struct {
	grpc.ServerStream
}

func (x *objectMigrationReceiveDataServer) SendAndClose(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

func (x *objectMigrationReceiveDataServer) Recv() (*ObjectData, error) {
	m := new(ObjectData)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _ObjectMigration_FinishReceive_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ObjectInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectMigrationServer).FinishReceive(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ObjectMigration/FinishReceive",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectMigrationServer).FinishReceive(ctx, req.(*ObjectInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _ObjectMigration_RequireSend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequiredInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectMigrationServer).RequireSend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ObjectMigration/RequireSend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectMigrationServer).RequireSend(ctx, req.(*RequiredInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _ObjectMigration_LeaveCommand_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectMigrationServer).LeaveCommand(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ObjectMigration/LeaveCommand",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectMigrationServer).LeaveCommand(ctx, req.(*EmptyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ObjectMigration_JoinCommand_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectMigrationServer).JoinCommand(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ObjectMigration/JoinCommand",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectMigrationServer).JoinCommand(ctx, req.(*EmptyReq))
	}
	return interceptor(ctx, in, info, handler)
}

// ObjectMigration_ServiceDesc is the grpc.ServiceDesc for ObjectMigration service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ObjectMigration_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.ObjectMigration",
	HandlerType: (*ObjectMigrationServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FinishReceive",
			Handler:    _ObjectMigration_FinishReceive_Handler,
		},
		{
			MethodName: "RequireSend",
			Handler:    _ObjectMigration_RequireSend_Handler,
		},
		{
			MethodName: "LeaveCommand",
			Handler:    _ObjectMigration_LeaveCommand_Handler,
		},
		{
			MethodName: "JoinCommand",
			Handler:    _ObjectMigration_JoinCommand_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ReceiveData",
			Handler:       _ObjectMigration_ReceiveData_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "object_migration.proto",
}
