// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package net

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// VoterQueryClient is the client API for VoterQuery service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type VoterQueryClient interface {
	Opinion(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (*QueryReply, error)
}

type voterQueryClient struct {
	cc grpc.ClientConnInterface
}

func NewVoterQueryClient(cc grpc.ClientConnInterface) VoterQueryClient {
	return &voterQueryClient{cc}
}

func (c *voterQueryClient) Opinion(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (*QueryReply, error) {
	out := new(QueryReply)
	err := c.cc.Invoke(ctx, "/net.VoterQuery/Opinion", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VoterQueryServer is the server API for VoterQuery service.
// All implementations must embed UnimplementedVoterQueryServer
// for forward compatibility
type VoterQueryServer interface {
	Opinion(context.Context, *QueryRequest) (*QueryReply, error)
	mustEmbedUnimplementedVoterQueryServer()
}

// UnimplementedVoterQueryServer must be embedded to have forward compatible implementations.
type UnimplementedVoterQueryServer struct {
}

func (UnimplementedVoterQueryServer) Opinion(context.Context, *QueryRequest) (*QueryReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Opinion not implemented")
}
func (UnimplementedVoterQueryServer) mustEmbedUnimplementedVoterQueryServer() {}

// UnsafeVoterQueryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to VoterQueryServer will
// result in compilation errors.
type UnsafeVoterQueryServer interface {
	mustEmbedUnimplementedVoterQueryServer()
}

func RegisterVoterQueryServer(s grpc.ServiceRegistrar, srv VoterQueryServer) {
	s.RegisterService(&_VoterQuery_serviceDesc, srv)
}

func _VoterQuery_Opinion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VoterQueryServer).Opinion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/net.VoterQuery/Opinion",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VoterQueryServer).Opinion(ctx, req.(*QueryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _VoterQuery_serviceDesc = grpc.ServiceDesc{
	ServiceName: "net.VoterQuery",
	HandlerType: (*VoterQueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Opinion",
			Handler:    _VoterQuery_Opinion_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "packages/vote/net/query.proto",
}