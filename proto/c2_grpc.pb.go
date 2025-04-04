// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v6.30.1
// source: proto/c2.proto

package proto

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

// C2ServiceClient is the client API for C2Service service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type C2ServiceClient interface {
	// Command stream from server to client
	SendCommands(ctx context.Context, opts ...grpc.CallOption) (C2Service_SendCommandsClient, error)
	// Client registration
	RegisterAgent(ctx context.Context, in *AgentInfo, opts ...grpc.CallOption) (*RegistrationResponse, error)
	// Client heartbeat
	Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error)
}

type c2ServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewC2ServiceClient(cc grpc.ClientConnInterface) C2ServiceClient {
	return &c2ServiceClient{cc}
}

func (c *c2ServiceClient) SendCommands(ctx context.Context, opts ...grpc.CallOption) (C2Service_SendCommandsClient, error) {
	stream, err := c.cc.NewStream(ctx, &C2Service_ServiceDesc.Streams[0], "/c2.C2Service/SendCommands", opts...)
	if err != nil {
		return nil, err
	}
	x := &c2ServiceSendCommandsClient{stream}
	return x, nil
}

type C2Service_SendCommandsClient interface {
	Send(*Command) error
	Recv() (*CommandResponse, error)
	grpc.ClientStream
}

type c2ServiceSendCommandsClient struct {
	grpc.ClientStream
}

func (x *c2ServiceSendCommandsClient) Send(m *Command) error {
	return x.ClientStream.SendMsg(m)
}

func (x *c2ServiceSendCommandsClient) Recv() (*CommandResponse, error) {
	m := new(CommandResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *c2ServiceClient) RegisterAgent(ctx context.Context, in *AgentInfo, opts ...grpc.CallOption) (*RegistrationResponse, error) {
	out := new(RegistrationResponse)
	err := c.cc.Invoke(ctx, "/c2.C2Service/RegisterAgent", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *c2ServiceClient) Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error) {
	out := new(HeartbeatResponse)
	err := c.cc.Invoke(ctx, "/c2.C2Service/Heartbeat", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// C2ServiceServer is the server API for C2Service service.
// All implementations must embed UnimplementedC2ServiceServer
// for forward compatibility
type C2ServiceServer interface {
	// Command stream from server to client
	SendCommands(C2Service_SendCommandsServer) error
	// Client registration
	RegisterAgent(context.Context, *AgentInfo) (*RegistrationResponse, error)
	// Client heartbeat
	Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error)
	mustEmbedUnimplementedC2ServiceServer()
}

// UnimplementedC2ServiceServer must be embedded to have forward compatible implementations.
type UnimplementedC2ServiceServer struct {
}

func (UnimplementedC2ServiceServer) SendCommands(C2Service_SendCommandsServer) error {
	return status.Errorf(codes.Unimplemented, "method SendCommands not implemented")
}
func (UnimplementedC2ServiceServer) RegisterAgent(context.Context, *AgentInfo) (*RegistrationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterAgent not implemented")
}
func (UnimplementedC2ServiceServer) Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Heartbeat not implemented")
}
func (UnimplementedC2ServiceServer) mustEmbedUnimplementedC2ServiceServer() {}

// UnsafeC2ServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to C2ServiceServer will
// result in compilation errors.
type UnsafeC2ServiceServer interface {
	mustEmbedUnimplementedC2ServiceServer()
}

func RegisterC2ServiceServer(s grpc.ServiceRegistrar, srv C2ServiceServer) {
	s.RegisterService(&C2Service_ServiceDesc, srv)
}

func _C2Service_SendCommands_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(C2ServiceServer).SendCommands(&c2ServiceSendCommandsServer{stream})
}

type C2Service_SendCommandsServer interface {
	Send(*CommandResponse) error
	Recv() (*Command, error)
	grpc.ServerStream
}

type c2ServiceSendCommandsServer struct {
	grpc.ServerStream
}

func (x *c2ServiceSendCommandsServer) Send(m *CommandResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *c2ServiceSendCommandsServer) Recv() (*Command, error) {
	m := new(Command)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _C2Service_RegisterAgent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AgentInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(C2ServiceServer).RegisterAgent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/c2.C2Service/RegisterAgent",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(C2ServiceServer).RegisterAgent(ctx, req.(*AgentInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _C2Service_Heartbeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartbeatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(C2ServiceServer).Heartbeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/c2.C2Service/Heartbeat",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(C2ServiceServer).Heartbeat(ctx, req.(*HeartbeatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// C2Service_ServiceDesc is the grpc.ServiceDesc for C2Service service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var C2Service_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "c2.C2Service",
	HandlerType: (*C2ServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterAgent",
			Handler:    _C2Service_RegisterAgent_Handler,
		},
		{
			MethodName: "Heartbeat",
			Handler:    _C2Service_Heartbeat_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SendCommands",
			Handler:       _C2Service_SendCommands_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "proto/c2.proto",
}
