// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: staking/v1beta/querier.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types/query"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	grpc "google.golang.org/grpc"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

func init() { proto.RegisterFile("staking/v1beta/querier.proto", fileDescriptor_71279347500d57ae) }

var fileDescriptor_71279347500d57ae = []byte{
	// 181 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x34, 0xce, 0x3f, 0x0e, 0x82, 0x30,
	0x14, 0xc7, 0x71, 0x18, 0xd4, 0x84, 0xd1, 0xc9, 0x18, 0xd3, 0x03, 0x38, 0xf4, 0x89, 0xde, 0xc0,
	0xd1, 0xcd, 0xd5, 0xad, 0xc5, 0x06, 0x1a, 0x68, 0x1f, 0xb6, 0x0f, 0x95, 0x5b, 0x78, 0x2c, 0x47,
	0x46, 0x47, 0x03, 0x17, 0x31, 0xf2, 0x67, 0xff, 0xfe, 0x3e, 0xf9, 0x45, 0x1b, 0x4f, 0x22, 0xd7,
	0x36, 0x85, 0x7b, 0x2c, 0x15, 0x09, 0xb8, 0x55, 0xca, 0x69, 0xe5, 0x78, 0xe9, 0x90, 0x70, 0xb9,
	0xca, 0x94, 0x36, 0x57, 0x51, 0x14, 0x7c, 0xcc, 0xf8, 0x90, 0xc5, 0xeb, 0x6d, 0x82, 0xde, 0xa0,
	0x07, 0x29, 0xbc, 0xea, 0x47, 0xf5, 0x28, 0xc4, 0x50, 0x8a, 0x54, 0x5b, 0x41, 0x1a, 0xed, 0xa0,
	0xec, 0x17, 0xd1, 0xec, 0xfc, 0x2f, 0x8e, 0xa7, 0x77, 0xcb, 0xc2, 0xa6, 0x65, 0xe1, 0xb7, 0x65,
	0xe1, 0xab, 0x63, 0x41, 0xd3, 0xb1, 0xe0, 0xd3, 0xb1, 0xe0, 0xb2, 0x4b, 0x35, 0x65, 0x95, 0xe4,
	0x09, 0x1a, 0x30, 0x82, 0x74, 0x62, 0x15, 0x3d, 0xd0, 0xe5, 0x30, 0x1d, 0x80, 0x27, 0x4c, 0x4f,
	0xa9, 0x2e, 0x95, 0x97, 0xf3, 0xde, 0x3e, 0xfc, 0x02, 0x00, 0x00, 0xff, 0xff, 0x09, 0xbe, 0x5d,
	0x64, 0xc1, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

// QueryServer is the server API for Query service.
type QueryServer interface {
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "heimdall.staking.v1beta1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "staking/v1beta/querier.proto",
}