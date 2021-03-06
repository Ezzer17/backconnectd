// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package backconnectd

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

// BackconnectdClient is the client API for Backconnectd service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BackconnectdClient interface {
	KillSession(ctx context.Context, in *SessionKillRequest, opts ...grpc.CallOption) (*Response, error)
	ConnectSession(ctx context.Context, opts ...grpc.CallOption) (Backconnectd_ConnectSessionClient, error)
	Subscribe(ctx context.Context, in *SubscribeRequest, opts ...grpc.CallOption) (Backconnectd_SubscribeClient, error)
}

type backconnectdClient struct {
	cc grpc.ClientConnInterface
}

func NewBackconnectdClient(cc grpc.ClientConnInterface) BackconnectdClient {
	return &backconnectdClient{cc}
}

func (c *backconnectdClient) KillSession(ctx context.Context, in *SessionKillRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/grpc.Backconnectd/KillSession", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *backconnectdClient) ConnectSession(ctx context.Context, opts ...grpc.CallOption) (Backconnectd_ConnectSessionClient, error) {
	stream, err := c.cc.NewStream(ctx, &Backconnectd_ServiceDesc.Streams[0], "/grpc.Backconnectd/ConnectSession", opts...)
	if err != nil {
		return nil, err
	}
	x := &backconnectdConnectSessionClient{stream}
	return x, nil
}

type Backconnectd_ConnectSessionClient interface {
	Send(*RawData) error
	Recv() (*RawData, error)
	grpc.ClientStream
}

type backconnectdConnectSessionClient struct {
	grpc.ClientStream
}

func (x *backconnectdConnectSessionClient) Send(m *RawData) error {
	return x.ClientStream.SendMsg(m)
}

func (x *backconnectdConnectSessionClient) Recv() (*RawData, error) {
	m := new(RawData)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *backconnectdClient) Subscribe(ctx context.Context, in *SubscribeRequest, opts ...grpc.CallOption) (Backconnectd_SubscribeClient, error) {
	stream, err := c.cc.NewStream(ctx, &Backconnectd_ServiceDesc.Streams[1], "/grpc.Backconnectd/Subscribe", opts...)
	if err != nil {
		return nil, err
	}
	x := &backconnectdSubscribeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Backconnectd_SubscribeClient interface {
	Recv() (*SessionEvent, error)
	grpc.ClientStream
}

type backconnectdSubscribeClient struct {
	grpc.ClientStream
}

func (x *backconnectdSubscribeClient) Recv() (*SessionEvent, error) {
	m := new(SessionEvent)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// BackconnectdServer is the server API for Backconnectd service.
// All implementations must embed UnimplementedBackconnectdServer
// for forward compatibility
type BackconnectdServer interface {
	KillSession(context.Context, *SessionKillRequest) (*Response, error)
	ConnectSession(Backconnectd_ConnectSessionServer) error
	Subscribe(*SubscribeRequest, Backconnectd_SubscribeServer) error
	mustEmbedUnimplementedBackconnectdServer()
}

// UnimplementedBackconnectdServer must be embedded to have forward compatible implementations.
type UnimplementedBackconnectdServer struct {
}

func (UnimplementedBackconnectdServer) KillSession(context.Context, *SessionKillRequest) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method KillSession not implemented")
}
func (UnimplementedBackconnectdServer) ConnectSession(Backconnectd_ConnectSessionServer) error {
	return status.Errorf(codes.Unimplemented, "method ConnectSession not implemented")
}
func (UnimplementedBackconnectdServer) Subscribe(*SubscribeRequest, Backconnectd_SubscribeServer) error {
	return status.Errorf(codes.Unimplemented, "method Subscribe not implemented")
}
func (UnimplementedBackconnectdServer) mustEmbedUnimplementedBackconnectdServer() {}

// UnsafeBackconnectdServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BackconnectdServer will
// result in compilation errors.
type UnsafeBackconnectdServer interface {
	mustEmbedUnimplementedBackconnectdServer()
}

func RegisterBackconnectdServer(s grpc.ServiceRegistrar, srv BackconnectdServer) {
	s.RegisterService(&Backconnectd_ServiceDesc, srv)
}

func _Backconnectd_KillSession_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SessionKillRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BackconnectdServer).KillSession(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.Backconnectd/KillSession",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BackconnectdServer).KillSession(ctx, req.(*SessionKillRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Backconnectd_ConnectSession_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(BackconnectdServer).ConnectSession(&backconnectdConnectSessionServer{stream})
}

type Backconnectd_ConnectSessionServer interface {
	Send(*RawData) error
	Recv() (*RawData, error)
	grpc.ServerStream
}

type backconnectdConnectSessionServer struct {
	grpc.ServerStream
}

func (x *backconnectdConnectSessionServer) Send(m *RawData) error {
	return x.ServerStream.SendMsg(m)
}

func (x *backconnectdConnectSessionServer) Recv() (*RawData, error) {
	m := new(RawData)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Backconnectd_Subscribe_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SubscribeRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(BackconnectdServer).Subscribe(m, &backconnectdSubscribeServer{stream})
}

type Backconnectd_SubscribeServer interface {
	Send(*SessionEvent) error
	grpc.ServerStream
}

type backconnectdSubscribeServer struct {
	grpc.ServerStream
}

func (x *backconnectdSubscribeServer) Send(m *SessionEvent) error {
	return x.ServerStream.SendMsg(m)
}

// Backconnectd_ServiceDesc is the grpc.ServiceDesc for Backconnectd service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Backconnectd_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.Backconnectd",
	HandlerType: (*BackconnectdServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "KillSession",
			Handler:    _Backconnectd_KillSession_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ConnectSession",
			Handler:       _Backconnectd_ConnectSession_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "Subscribe",
			Handler:       _Backconnectd_Subscribe_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "service.proto",
}
