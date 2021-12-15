// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package chatterbox

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ChatterBoxClient is the client API for ChatterBox service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChatterBoxClient interface {
	// Chat joins the chat room and sends chat messages.
	Chat(ctx context.Context, opts ...grpc.CallOption) (ChatterBox_ChatClient, error)
	// Monitor passively monitors the room.
	Monitor(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (ChatterBox_MonitorClient, error)
}

type chatterBoxClient struct {
	cc grpc.ClientConnInterface
}

func NewChatterBoxClient(cc grpc.ClientConnInterface) ChatterBoxClient {
	return &chatterBoxClient{cc}
}

func (c *chatterBoxClient) Chat(ctx context.Context, opts ...grpc.CallOption) (ChatterBox_ChatClient, error) {
	stream, err := c.cc.NewStream(ctx, &ChatterBox_ServiceDesc.Streams[0], "/chatterbox.ChatterBox/Chat", opts...)
	if err != nil {
		return nil, err
	}
	x := &chatterBoxChatClient{stream}
	return x, nil
}

type ChatterBox_ChatClient interface {
	Send(*Send) error
	Recv() (*Event, error)
	grpc.ClientStream
}

type chatterBoxChatClient struct {
	grpc.ClientStream
}

func (x *chatterBoxChatClient) Send(m *Send) error {
	return x.ClientStream.SendMsg(m)
}

func (x *chatterBoxChatClient) Recv() (*Event, error) {
	m := new(Event)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *chatterBoxClient) Monitor(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (ChatterBox_MonitorClient, error) {
	stream, err := c.cc.NewStream(ctx, &ChatterBox_ServiceDesc.Streams[1], "/chatterbox.ChatterBox/Monitor", opts...)
	if err != nil {
		return nil, err
	}
	x := &chatterBoxMonitorClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ChatterBox_MonitorClient interface {
	Recv() (*Event, error)
	grpc.ClientStream
}

type chatterBoxMonitorClient struct {
	grpc.ClientStream
}

func (x *chatterBoxMonitorClient) Recv() (*Event, error) {
	m := new(Event)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ChatterBoxServer is the server API for ChatterBox service.
// All implementations must embed UnimplementedChatterBoxServer
// for forward compatibility
type ChatterBoxServer interface {
	// Chat joins the chat room and sends chat messages.
	Chat(ChatterBox_ChatServer) error
	// Monitor passively monitors the room.
	Monitor(*emptypb.Empty, ChatterBox_MonitorServer) error
	mustEmbedUnimplementedChatterBoxServer()
}

// UnimplementedChatterBoxServer must be embedded to have forward compatible implementations.
type UnimplementedChatterBoxServer struct {
}

func (UnimplementedChatterBoxServer) Chat(ChatterBox_ChatServer) error {
	return status.Errorf(codes.Unimplemented, "method Chat not implemented")
}
func (UnimplementedChatterBoxServer) Monitor(*emptypb.Empty, ChatterBox_MonitorServer) error {
	return status.Errorf(codes.Unimplemented, "method Monitor not implemented")
}
func (UnimplementedChatterBoxServer) mustEmbedUnimplementedChatterBoxServer() {}

// UnsafeChatterBoxServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChatterBoxServer will
// result in compilation errors.
type UnsafeChatterBoxServer interface {
	mustEmbedUnimplementedChatterBoxServer()
}

func RegisterChatterBoxServer(s grpc.ServiceRegistrar, srv ChatterBoxServer) {
	s.RegisterService(&ChatterBox_ServiceDesc, srv)
}

func _ChatterBox_Chat_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ChatterBoxServer).Chat(&chatterBoxChatServer{stream})
}

type ChatterBox_ChatServer interface {
	Send(*Event) error
	Recv() (*Send, error)
	grpc.ServerStream
}

type chatterBoxChatServer struct {
	grpc.ServerStream
}

func (x *chatterBoxChatServer) Send(m *Event) error {
	return x.ServerStream.SendMsg(m)
}

func (x *chatterBoxChatServer) Recv() (*Send, error) {
	m := new(Send)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _ChatterBox_Monitor_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(emptypb.Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ChatterBoxServer).Monitor(m, &chatterBoxMonitorServer{stream})
}

type ChatterBox_MonitorServer interface {
	Send(*Event) error
	grpc.ServerStream
}

type chatterBoxMonitorServer struct {
	grpc.ServerStream
}

func (x *chatterBoxMonitorServer) Send(m *Event) error {
	return x.ServerStream.SendMsg(m)
}

// ChatterBox_ServiceDesc is the grpc.ServiceDesc for ChatterBox service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ChatterBox_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "chatterbox.ChatterBox",
	HandlerType: (*ChatterBoxServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Chat",
			Handler:       _ChatterBox_Chat_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "Monitor",
			Handler:       _ChatterBox_Monitor_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "chatterbox.proto",
}
