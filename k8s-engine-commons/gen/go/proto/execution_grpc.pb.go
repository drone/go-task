// Copyright 2025 Harness Inc. All rights reserved.
// Use of this source code is governed by the PolyForm Free Trial 1.0.0 license
// that can be found in the licenses directory at the root of this repository, also available at
// https://polyformproject.org/wp-content/uploads/2020/05/PolyForm-Free-Trial-1.0.0.txt.

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: proto/execution.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	LiteEngine_UpdateState_FullMethodName        = "/io.harness.product.ci.engine.proto.LiteEngine/UpdateState"
	LiteEngine_GetImageEntrypoint_FullMethodName = "/io.harness.product.ci.engine.proto.LiteEngine/GetImageEntrypoint"
	LiteEngine_EvaluateJEXL_FullMethodName       = "/io.harness.product.ci.engine.proto.LiteEngine/EvaluateJEXL"
	LiteEngine_Ping_FullMethodName               = "/io.harness.product.ci.engine.proto.LiteEngine/Ping"
	LiteEngine_ExecuteStep_FullMethodName        = "/io.harness.product.ci.engine.proto.LiteEngine/ExecuteStep"
)

// LiteEngineClient is the client API for LiteEngine service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LiteEngineClient interface {
	// Synchronous RPC to execute a step
	// Deprecated
	UpdateState(ctx context.Context, in *UpdateStateRequest, opts ...grpc.CallOption) (*UpdateStateResponse, error)
	// Synchronous RPC to fetch image entrypoint
	GetImageEntrypoint(ctx context.Context, in *GetImageEntrypointRequest, opts ...grpc.CallOption) (*GetImageEntrypointResponse, error)
	// Synchronous RPC to evaluate JEXL expression
	EvaluateJEXL(ctx context.Context, in *EvaluateJEXLRequest, opts ...grpc.CallOption) (*EvaluateJEXLResponse, error)
	// Synchronous RPC to check health of lite-engine service.
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	// Asynchronous RPC that starts execution of a step.
	// It is idempotent such that if two requests are fired with same id, then
	// only one request will start execution of the step.
	ExecuteStep(ctx context.Context, in *ExecuteStepRequest, opts ...grpc.CallOption) (*ExecuteStepResponse, error)
}

type liteEngineClient struct {
	cc grpc.ClientConnInterface
}

func NewLiteEngineClient(cc grpc.ClientConnInterface) LiteEngineClient {
	return &liteEngineClient{cc}
}

func (c *liteEngineClient) UpdateState(ctx context.Context, in *UpdateStateRequest, opts ...grpc.CallOption) (*UpdateStateResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateStateResponse)
	err := c.cc.Invoke(ctx, LiteEngine_UpdateState_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *liteEngineClient) GetImageEntrypoint(ctx context.Context, in *GetImageEntrypointRequest, opts ...grpc.CallOption) (*GetImageEntrypointResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetImageEntrypointResponse)
	err := c.cc.Invoke(ctx, LiteEngine_GetImageEntrypoint_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *liteEngineClient) EvaluateJEXL(ctx context.Context, in *EvaluateJEXLRequest, opts ...grpc.CallOption) (*EvaluateJEXLResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EvaluateJEXLResponse)
	err := c.cc.Invoke(ctx, LiteEngine_EvaluateJEXL_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *liteEngineClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, LiteEngine_Ping_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *liteEngineClient) ExecuteStep(ctx context.Context, in *ExecuteStepRequest, opts ...grpc.CallOption) (*ExecuteStepResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ExecuteStepResponse)
	err := c.cc.Invoke(ctx, LiteEngine_ExecuteStep_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LiteEngineServer is the server API for LiteEngine service.
// All implementations must embed UnimplementedLiteEngineServer
// for forward compatibility.
type LiteEngineServer interface {
	// Synchronous RPC to execute a step
	// Deprecated
	UpdateState(context.Context, *UpdateStateRequest) (*UpdateStateResponse, error)
	// Synchronous RPC to fetch image entrypoint
	GetImageEntrypoint(context.Context, *GetImageEntrypointRequest) (*GetImageEntrypointResponse, error)
	// Synchronous RPC to evaluate JEXL expression
	EvaluateJEXL(context.Context, *EvaluateJEXLRequest) (*EvaluateJEXLResponse, error)
	// Synchronous RPC to check health of lite-engine service.
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	// Asynchronous RPC that starts execution of a step.
	// It is idempotent such that if two requests are fired with same id, then
	// only one request will start execution of the step.
	ExecuteStep(context.Context, *ExecuteStepRequest) (*ExecuteStepResponse, error)
	mustEmbedUnimplementedLiteEngineServer()
}

// UnimplementedLiteEngineServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedLiteEngineServer struct{}

func (UnimplementedLiteEngineServer) UpdateState(context.Context, *UpdateStateRequest) (*UpdateStateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateState not implemented")
}
func (UnimplementedLiteEngineServer) GetImageEntrypoint(context.Context, *GetImageEntrypointRequest) (*GetImageEntrypointResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetImageEntrypoint not implemented")
}
func (UnimplementedLiteEngineServer) EvaluateJEXL(context.Context, *EvaluateJEXLRequest) (*EvaluateJEXLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EvaluateJEXL not implemented")
}
func (UnimplementedLiteEngineServer) Ping(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedLiteEngineServer) ExecuteStep(context.Context, *ExecuteStepRequest) (*ExecuteStepResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecuteStep not implemented")
}
func (UnimplementedLiteEngineServer) mustEmbedUnimplementedLiteEngineServer() {}
func (UnimplementedLiteEngineServer) testEmbeddedByValue()                    {}

// UnsafeLiteEngineServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LiteEngineServer will
// result in compilation errors.
type UnsafeLiteEngineServer interface {
	mustEmbedUnimplementedLiteEngineServer()
}

func RegisterLiteEngineServer(s grpc.ServiceRegistrar, srv LiteEngineServer) {
	// If the following call pancis, it indicates UnimplementedLiteEngineServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&LiteEngine_ServiceDesc, srv)
}

func _LiteEngine_UpdateState_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateStateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LiteEngineServer).UpdateState(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LiteEngine_UpdateState_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LiteEngineServer).UpdateState(ctx, req.(*UpdateStateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LiteEngine_GetImageEntrypoint_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetImageEntrypointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LiteEngineServer).GetImageEntrypoint(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LiteEngine_GetImageEntrypoint_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LiteEngineServer).GetImageEntrypoint(ctx, req.(*GetImageEntrypointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LiteEngine_EvaluateJEXL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EvaluateJEXLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LiteEngineServer).EvaluateJEXL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LiteEngine_EvaluateJEXL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LiteEngineServer).EvaluateJEXL(ctx, req.(*EvaluateJEXLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LiteEngine_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LiteEngineServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LiteEngine_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LiteEngineServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LiteEngine_ExecuteStep_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExecuteStepRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LiteEngineServer).ExecuteStep(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LiteEngine_ExecuteStep_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LiteEngineServer).ExecuteStep(ctx, req.(*ExecuteStepRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// LiteEngine_ServiceDesc is the grpc.ServiceDesc for LiteEngine service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var LiteEngine_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "io.harness.product.ci.engine.proto.LiteEngine",
	HandlerType: (*LiteEngineServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "UpdateState",
			Handler:    _LiteEngine_UpdateState_Handler,
		},
		{
			MethodName: "GetImageEntrypoint",
			Handler:    _LiteEngine_GetImageEntrypoint_Handler,
		},
		{
			MethodName: "EvaluateJEXL",
			Handler:    _LiteEngine_EvaluateJEXL_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _LiteEngine_Ping_Handler,
		},
		{
			MethodName: "ExecuteStep",
			Handler:    _LiteEngine_ExecuteStep_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/execution.proto",
}
