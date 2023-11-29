// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: kyverdk/runtime/v1/runtime.proto

package v1

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

const (
	RuntimeService_GetRuntimeName_FullMethodName      = "/kyverdk.runtime.v1.RuntimeService/GetRuntimeName"
	RuntimeService_GetRuntimeVersion_FullMethodName   = "/kyverdk.runtime.v1.RuntimeService/GetRuntimeVersion"
	RuntimeService_ValidateSetConfig_FullMethodName   = "/kyverdk.runtime.v1.RuntimeService/ValidateSetConfig"
	RuntimeService_GetDataItem_FullMethodName         = "/kyverdk.runtime.v1.RuntimeService/GetDataItem"
	RuntimeService_PrevalidateDataItem_FullMethodName = "/kyverdk.runtime.v1.RuntimeService/PrevalidateDataItem"
	RuntimeService_TransformDataItem_FullMethodName   = "/kyverdk.runtime.v1.RuntimeService/TransformDataItem"
	RuntimeService_ValidateDataItem_FullMethodName    = "/kyverdk.runtime.v1.RuntimeService/ValidateDataItem"
	RuntimeService_SummarizeDataBundle_FullMethodName = "/kyverdk.runtime.v1.RuntimeService/SummarizeDataBundle"
	RuntimeService_NextKey_FullMethodName             = "/kyverdk.runtime.v1.RuntimeService/NextKey"
)

// RuntimeServiceClient is the client API for RuntimeService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RuntimeServiceClient interface {
	// Returns the name of the runtime. Example "@kyvejs/tendermint"
	GetRuntimeName(ctx context.Context, in *GetRuntimeNameRequest, opts ...grpc.CallOption) (*GetRuntimeNameResponse, error)
	// Returns the version of the runtime. Example "1.2.0"
	GetRuntimeVersion(ctx context.Context, in *GetRuntimeVersionRequest, opts ...grpc.CallOption) (*GetRuntimeVersionResponse, error)
	// Parses the raw runtime config found on pool, validates it and finally sets
	// the property "config" in the runtime. A raw config could be an ipfs link to the
	// actual config or a stringified yaml or json string. This method should error if
	// the specific runtime config is not parsable or invalid.
	//
	// Deterministic behavior is required
	ValidateSetConfig(ctx context.Context, in *ValidateSetConfigRequest, opts ...grpc.CallOption) (*ValidateSetConfigResponse, error)
	// Gets the data item from a specific key and returns both key and the value.
	//
	// Deterministic behavior is required
	GetDataItem(ctx context.Context, in *GetDataItemRequest, opts ...grpc.CallOption) (*GetDataItemResponse, error)
	// Prevalidates a data item right after is was retrieved from source.
	// If the prevalidation fails the item gets rejected and never makes
	// it to the local cache. If the prevalidation succeeds the item gets
	// transformed and written to cache were it is used from submission
	// of proposals or bundle validation.
	//
	// Deterministic behavior is required
	PrevalidateDataItem(ctx context.Context, in *PrevalidateDataItemRequest, opts ...grpc.CallOption) (*PrevalidateDataItemResponse, error)
	// Transforms a single data item and return it. Used for example
	// to remove unecessary data or format the data in a better way.
	//
	// Deterministic behavior is required
	TransformDataItem(ctx context.Context, in *TransformDataItemRequest, opts ...grpc.CallOption) (*TransformDataItemResponse, error)
	// Validates a single data item of a bundle proposal
	//
	// Deterministic behavior is required
	ValidateDataItem(ctx context.Context, in *ValidateDataItemRequest, opts ...grpc.CallOption) (*ValidateDataItemResponse, error)
	// Gets a formatted value string from a bundle. This produces a "summary" of
	// a bundle which gets stored on-chain and therefore needs to be short.
	//
	// String should not be longer than 100 characters, else gas costs might be too expensive.
	//
	// Deterministic behavior is required
	SummarizeDataBundle(ctx context.Context, in *SummarizeDataBundleRequest, opts ...grpc.CallOption) (*SummarizeDataBundleResponse, error)
	// Gets the next key from the current key so that the data archived has an order.
	//
	// Deterministic behavior is required
	NextKey(ctx context.Context, in *NextKeyRequest, opts ...grpc.CallOption) (*NextKeyResponse, error)
}

type runtimeServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRuntimeServiceClient(cc grpc.ClientConnInterface) RuntimeServiceClient {
	return &runtimeServiceClient{cc}
}

func (c *runtimeServiceClient) GetRuntimeName(ctx context.Context, in *GetRuntimeNameRequest, opts ...grpc.CallOption) (*GetRuntimeNameResponse, error) {
	out := new(GetRuntimeNameResponse)
	err := c.cc.Invoke(ctx, RuntimeService_GetRuntimeName_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) GetRuntimeVersion(ctx context.Context, in *GetRuntimeVersionRequest, opts ...grpc.CallOption) (*GetRuntimeVersionResponse, error) {
	out := new(GetRuntimeVersionResponse)
	err := c.cc.Invoke(ctx, RuntimeService_GetRuntimeVersion_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) ValidateSetConfig(ctx context.Context, in *ValidateSetConfigRequest, opts ...grpc.CallOption) (*ValidateSetConfigResponse, error) {
	out := new(ValidateSetConfigResponse)
	err := c.cc.Invoke(ctx, RuntimeService_ValidateSetConfig_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) GetDataItem(ctx context.Context, in *GetDataItemRequest, opts ...grpc.CallOption) (*GetDataItemResponse, error) {
	out := new(GetDataItemResponse)
	err := c.cc.Invoke(ctx, RuntimeService_GetDataItem_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) PrevalidateDataItem(ctx context.Context, in *PrevalidateDataItemRequest, opts ...grpc.CallOption) (*PrevalidateDataItemResponse, error) {
	out := new(PrevalidateDataItemResponse)
	err := c.cc.Invoke(ctx, RuntimeService_PrevalidateDataItem_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) TransformDataItem(ctx context.Context, in *TransformDataItemRequest, opts ...grpc.CallOption) (*TransformDataItemResponse, error) {
	out := new(TransformDataItemResponse)
	err := c.cc.Invoke(ctx, RuntimeService_TransformDataItem_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) ValidateDataItem(ctx context.Context, in *ValidateDataItemRequest, opts ...grpc.CallOption) (*ValidateDataItemResponse, error) {
	out := new(ValidateDataItemResponse)
	err := c.cc.Invoke(ctx, RuntimeService_ValidateDataItem_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) SummarizeDataBundle(ctx context.Context, in *SummarizeDataBundleRequest, opts ...grpc.CallOption) (*SummarizeDataBundleResponse, error) {
	out := new(SummarizeDataBundleResponse)
	err := c.cc.Invoke(ctx, RuntimeService_SummarizeDataBundle_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runtimeServiceClient) NextKey(ctx context.Context, in *NextKeyRequest, opts ...grpc.CallOption) (*NextKeyResponse, error) {
	out := new(NextKeyResponse)
	err := c.cc.Invoke(ctx, RuntimeService_NextKey_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RuntimeServiceServer is the server API for RuntimeService service.
// All implementations must embed UnimplementedRuntimeServiceServer
// for forward compatibility
type RuntimeServiceServer interface {
	// Returns the name of the runtime. Example "@kyvejs/tendermint"
	GetRuntimeName(context.Context, *GetRuntimeNameRequest) (*GetRuntimeNameResponse, error)
	// Returns the version of the runtime. Example "1.2.0"
	GetRuntimeVersion(context.Context, *GetRuntimeVersionRequest) (*GetRuntimeVersionResponse, error)
	// Parses the raw runtime config found on pool, validates it and finally sets
	// the property "config" in the runtime. A raw config could be an ipfs link to the
	// actual config or a stringified yaml or json string. This method should error if
	// the specific runtime config is not parsable or invalid.
	//
	// Deterministic behavior is required
	ValidateSetConfig(context.Context, *ValidateSetConfigRequest) (*ValidateSetConfigResponse, error)
	// Gets the data item from a specific key and returns both key and the value.
	//
	// Deterministic behavior is required
	GetDataItem(context.Context, *GetDataItemRequest) (*GetDataItemResponse, error)
	// Prevalidates a data item right after is was retrieved from source.
	// If the prevalidation fails the item gets rejected and never makes
	// it to the local cache. If the prevalidation succeeds the item gets
	// transformed and written to cache were it is used from submission
	// of proposals or bundle validation.
	//
	// Deterministic behavior is required
	PrevalidateDataItem(context.Context, *PrevalidateDataItemRequest) (*PrevalidateDataItemResponse, error)
	// Transforms a single data item and return it. Used for example
	// to remove unecessary data or format the data in a better way.
	//
	// Deterministic behavior is required
	TransformDataItem(context.Context, *TransformDataItemRequest) (*TransformDataItemResponse, error)
	// Validates a single data item of a bundle proposal
	//
	// Deterministic behavior is required
	ValidateDataItem(context.Context, *ValidateDataItemRequest) (*ValidateDataItemResponse, error)
	// Gets a formatted value string from a bundle. This produces a "summary" of
	// a bundle which gets stored on-chain and therefore needs to be short.
	//
	// String should not be longer than 100 characters, else gas costs might be too expensive.
	//
	// Deterministic behavior is required
	SummarizeDataBundle(context.Context, *SummarizeDataBundleRequest) (*SummarizeDataBundleResponse, error)
	// Gets the next key from the current key so that the data archived has an order.
	//
	// Deterministic behavior is required
	NextKey(context.Context, *NextKeyRequest) (*NextKeyResponse, error)
	mustEmbedUnimplementedRuntimeServiceServer()
}

// UnimplementedRuntimeServiceServer must be embedded to have forward compatible implementations.
type UnimplementedRuntimeServiceServer struct {
}

func (UnimplementedRuntimeServiceServer) GetRuntimeName(context.Context, *GetRuntimeNameRequest) (*GetRuntimeNameResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRuntimeName not implemented")
}
func (UnimplementedRuntimeServiceServer) GetRuntimeVersion(context.Context, *GetRuntimeVersionRequest) (*GetRuntimeVersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRuntimeVersion not implemented")
}
func (UnimplementedRuntimeServiceServer) ValidateSetConfig(context.Context, *ValidateSetConfigRequest) (*ValidateSetConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateSetConfig not implemented")
}
func (UnimplementedRuntimeServiceServer) GetDataItem(context.Context, *GetDataItemRequest) (*GetDataItemResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDataItem not implemented")
}
func (UnimplementedRuntimeServiceServer) PrevalidateDataItem(context.Context, *PrevalidateDataItemRequest) (*PrevalidateDataItemResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PrevalidateDataItem not implemented")
}
func (UnimplementedRuntimeServiceServer) TransformDataItem(context.Context, *TransformDataItemRequest) (*TransformDataItemResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TransformDataItem not implemented")
}
func (UnimplementedRuntimeServiceServer) ValidateDataItem(context.Context, *ValidateDataItemRequest) (*ValidateDataItemResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateDataItem not implemented")
}
func (UnimplementedRuntimeServiceServer) SummarizeDataBundle(context.Context, *SummarizeDataBundleRequest) (*SummarizeDataBundleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SummarizeDataBundle not implemented")
}
func (UnimplementedRuntimeServiceServer) NextKey(context.Context, *NextKeyRequest) (*NextKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NextKey not implemented")
}
func (UnimplementedRuntimeServiceServer) mustEmbedUnimplementedRuntimeServiceServer() {}

// UnsafeRuntimeServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RuntimeServiceServer will
// result in compilation errors.
type UnsafeRuntimeServiceServer interface {
	mustEmbedUnimplementedRuntimeServiceServer()
}

func RegisterRuntimeServiceServer(s grpc.ServiceRegistrar, srv RuntimeServiceServer) {
	s.RegisterService(&RuntimeService_ServiceDesc, srv)
}

func _RuntimeService_GetRuntimeName_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRuntimeNameRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).GetRuntimeName(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RuntimeService_GetRuntimeName_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).GetRuntimeName(ctx, req.(*GetRuntimeNameRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_GetRuntimeVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRuntimeVersionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).GetRuntimeVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RuntimeService_GetRuntimeVersion_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).GetRuntimeVersion(ctx, req.(*GetRuntimeVersionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_ValidateSetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateSetConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).ValidateSetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RuntimeService_ValidateSetConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).ValidateSetConfig(ctx, req.(*ValidateSetConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_GetDataItem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDataItemRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).GetDataItem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RuntimeService_GetDataItem_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).GetDataItem(ctx, req.(*GetDataItemRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_PrevalidateDataItem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PrevalidateDataItemRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).PrevalidateDataItem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RuntimeService_PrevalidateDataItem_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).PrevalidateDataItem(ctx, req.(*PrevalidateDataItemRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_TransformDataItem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransformDataItemRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).TransformDataItem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RuntimeService_TransformDataItem_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).TransformDataItem(ctx, req.(*TransformDataItemRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_ValidateDataItem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateDataItemRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).ValidateDataItem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RuntimeService_ValidateDataItem_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).ValidateDataItem(ctx, req.(*ValidateDataItemRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_SummarizeDataBundle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SummarizeDataBundleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).SummarizeDataBundle(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RuntimeService_SummarizeDataBundle_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).SummarizeDataBundle(ctx, req.(*SummarizeDataBundleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuntimeService_NextKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NextKeyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuntimeServiceServer).NextKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RuntimeService_NextKey_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuntimeServiceServer).NextKey(ctx, req.(*NextKeyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RuntimeService_ServiceDesc is the grpc.ServiceDesc for RuntimeService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RuntimeService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kyverdk.runtime.v1.RuntimeService",
	HandlerType: (*RuntimeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetRuntimeName",
			Handler:    _RuntimeService_GetRuntimeName_Handler,
		},
		{
			MethodName: "GetRuntimeVersion",
			Handler:    _RuntimeService_GetRuntimeVersion_Handler,
		},
		{
			MethodName: "ValidateSetConfig",
			Handler:    _RuntimeService_ValidateSetConfig_Handler,
		},
		{
			MethodName: "GetDataItem",
			Handler:    _RuntimeService_GetDataItem_Handler,
		},
		{
			MethodName: "PrevalidateDataItem",
			Handler:    _RuntimeService_PrevalidateDataItem_Handler,
		},
		{
			MethodName: "TransformDataItem",
			Handler:    _RuntimeService_TransformDataItem_Handler,
		},
		{
			MethodName: "ValidateDataItem",
			Handler:    _RuntimeService_ValidateDataItem_Handler,
		},
		{
			MethodName: "SummarizeDataBundle",
			Handler:    _RuntimeService_SummarizeDataBundle_Handler,
		},
		{
			MethodName: "NextKey",
			Handler:    _RuntimeService_NextKey_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kyverdk/runtime/v1/runtime.proto",
}
