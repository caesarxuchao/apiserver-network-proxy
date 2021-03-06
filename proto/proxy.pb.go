/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proxy.proto

package proxy

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type DialRequest struct {
	// tcp or udp?
	Protocol string `protobuf:"bytes,1,opt,name=protocol,proto3" json:"protocol,omitempty"`
	// node:port
	Address              string   `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DialRequest) Reset()         { *m = DialRequest{} }
func (m *DialRequest) String() string { return proto.CompactTextString(m) }
func (*DialRequest) ProtoMessage()    {}
func (*DialRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_proxy_af9cd44112350187, []int{0}
}
func (m *DialRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DialRequest.Unmarshal(m, b)
}
func (m *DialRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DialRequest.Marshal(b, m, deterministic)
}
func (dst *DialRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DialRequest.Merge(dst, src)
}
func (m *DialRequest) XXX_Size() int {
	return xxx_messageInfo_DialRequest.Size(m)
}
func (m *DialRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DialRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DialRequest proto.InternalMessageInfo

func (m *DialRequest) GetProtocol() string {
	if m != nil {
		return m.Protocol
	}
	return ""
}

func (m *DialRequest) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type DialResponse struct {
	Error                string   `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	StreamID             int32    `protobuf:"varint,2,opt,name=streamID,proto3" json:"streamID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DialResponse) Reset()         { *m = DialResponse{} }
func (m *DialResponse) String() string { return proto.CompactTextString(m) }
func (*DialResponse) ProtoMessage()    {}
func (*DialResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_proxy_af9cd44112350187, []int{1}
}
func (m *DialResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DialResponse.Unmarshal(m, b)
}
func (m *DialResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DialResponse.Marshal(b, m, deterministic)
}
func (dst *DialResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DialResponse.Merge(dst, src)
}
func (m *DialResponse) XXX_Size() int {
	return xxx_messageInfo_DialResponse.Size(m)
}
func (m *DialResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_DialResponse.DiscardUnknown(m)
}

var xxx_messageInfo_DialResponse proto.InternalMessageInfo

func (m *DialResponse) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func (m *DialResponse) GetStreamID() int32 {
	if m != nil {
		return m.StreamID
	}
	return 0
}

type Data struct {
	// streamID to connect to
	StreamID int32 `protobuf:"varint,1,opt,name=streamID,proto3" json:"streamID,omitempty"`
	// error message if error happens
	Error string `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	// stream data
	Data                 []byte   `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Data) Reset()         { *m = Data{} }
func (m *Data) String() string { return proto.CompactTextString(m) }
func (*Data) ProtoMessage()    {}
func (*Data) Descriptor() ([]byte, []int) {
	return fileDescriptor_proxy_af9cd44112350187, []int{2}
}
func (m *Data) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Data.Unmarshal(m, b)
}
func (m *Data) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Data.Marshal(b, m, deterministic)
}
func (dst *Data) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Data.Merge(dst, src)
}
func (m *Data) XXX_Size() int {
	return xxx_messageInfo_Data.Size(m)
}
func (m *Data) XXX_DiscardUnknown() {
	xxx_messageInfo_Data.DiscardUnknown(m)
}

var xxx_messageInfo_Data proto.InternalMessageInfo

func (m *Data) GetStreamID() int32 {
	if m != nil {
		return m.StreamID
	}
	return 0
}

func (m *Data) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func (m *Data) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*DialRequest)(nil), "DialRequest")
	proto.RegisterType((*DialResponse)(nil), "DialResponse")
	proto.RegisterType((*Data)(nil), "Data")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ProxyServiceClient is the client API for ProxyService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ProxyServiceClient interface {
	// Dial a remote address and return a stream id if success
	Dial(ctx context.Context, in *DialRequest, opts ...grpc.CallOption) (*DialResponse, error)
	// Connect connects to a remote address by stream id, and establish
	// a bi-directional stream.
	Connect(ctx context.Context, opts ...grpc.CallOption) (ProxyService_ConnectClient, error)
}

type proxyServiceClient struct {
	cc *grpc.ClientConn
}

func NewProxyServiceClient(cc *grpc.ClientConn) ProxyServiceClient {
	return &proxyServiceClient{cc}
}

func (c *proxyServiceClient) Dial(ctx context.Context, in *DialRequest, opts ...grpc.CallOption) (*DialResponse, error) {
	out := new(DialResponse)
	err := c.cc.Invoke(ctx, "/ProxyService/Dial", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *proxyServiceClient) Connect(ctx context.Context, opts ...grpc.CallOption) (ProxyService_ConnectClient, error) {
	stream, err := c.cc.NewStream(ctx, &_ProxyService_serviceDesc.Streams[0], "/ProxyService/Connect", opts...)
	if err != nil {
		return nil, err
	}
	x := &proxyServiceConnectClient{stream}
	return x, nil
}

type ProxyService_ConnectClient interface {
	Send(*Data) error
	Recv() (*Data, error)
	grpc.ClientStream
}

type proxyServiceConnectClient struct {
	grpc.ClientStream
}

func (x *proxyServiceConnectClient) Send(m *Data) error {
	return x.ClientStream.SendMsg(m)
}

func (x *proxyServiceConnectClient) Recv() (*Data, error) {
	m := new(Data)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ProxyServiceServer is the server API for ProxyService service.
type ProxyServiceServer interface {
	// Dial a remote address and return a stream id if success
	Dial(context.Context, *DialRequest) (*DialResponse, error)
	// Connect connects to a remote address by stream id, and establish
	// a bi-directional stream.
	Connect(ProxyService_ConnectServer) error
}

func RegisterProxyServiceServer(s *grpc.Server, srv ProxyServiceServer) {
	s.RegisterService(&_ProxyService_serviceDesc, srv)
}

func _ProxyService_Dial_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DialRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProxyServiceServer).Dial(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ProxyService/Dial",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProxyServiceServer).Dial(ctx, req.(*DialRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProxyService_Connect_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ProxyServiceServer).Connect(&proxyServiceConnectServer{stream})
}

type ProxyService_ConnectServer interface {
	Send(*Data) error
	Recv() (*Data, error)
	grpc.ServerStream
}

type proxyServiceConnectServer struct {
	grpc.ServerStream
}

func (x *proxyServiceConnectServer) Send(m *Data) error {
	return x.ServerStream.SendMsg(m)
}

func (x *proxyServiceConnectServer) Recv() (*Data, error) {
	m := new(Data)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _ProxyService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "ProxyService",
	HandlerType: (*ProxyServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Dial",
			Handler:    _ProxyService_Dial_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Connect",
			Handler:       _ProxyService_Connect_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "proxy.proto",
}

func init() { proto.RegisterFile("proxy.proto", fileDescriptor_proxy_af9cd44112350187) }

var fileDescriptor_proxy_af9cd44112350187 = []byte{
	// 222 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x90, 0xc1, 0x4a, 0xc4, 0x30,
	0x10, 0x86, 0x37, 0x6b, 0xeb, 0xea, 0x6c, 0xbc, 0x0c, 0x1e, 0x42, 0x41, 0x58, 0x02, 0x42, 0x4f,
	0x41, 0xf4, 0x05, 0x84, 0xed, 0x45, 0xf0, 0x20, 0xd1, 0x17, 0x88, 0xed, 0x1c, 0x16, 0xd6, 0xa6,
	0x4e, 0xa2, 0xe8, 0xdb, 0x4b, 0x52, 0xad, 0xed, 0x29, 0xf9, 0x92, 0xe1, 0xe3, 0xff, 0x07, 0xb6,
	0x03, 0xfb, 0xaf, 0x6f, 0x33, 0xb0, 0x8f, 0x5e, 0xef, 0x61, 0xdb, 0x1c, 0xdc, 0xd1, 0xd2, 0xfb,
	0x07, 0x85, 0x88, 0x15, 0x9c, 0xe5, 0xf7, 0xd6, 0x1f, 0x95, 0xd8, 0x89, 0xfa, 0xdc, 0x4e, 0x8c,
	0x0a, 0x36, 0xae, 0xeb, 0x98, 0x42, 0x50, 0xeb, 0xfc, 0xf5, 0x87, 0xfa, 0x1e, 0xe4, 0x28, 0x09,
	0x83, 0xef, 0x03, 0xe1, 0x25, 0x94, 0xc4, 0xec, 0xf9, 0x57, 0x31, 0x42, 0x72, 0x87, 0xc8, 0xe4,
	0xde, 0x1e, 0x9a, 0x2c, 0x28, 0xed, 0xc4, 0xfa, 0x11, 0x8a, 0xc6, 0x45, 0xb7, 0x98, 0x11, 0xcb,
	0x99, 0x7f, 0xeb, 0x7a, 0x6e, 0x45, 0x28, 0x3a, 0x17, 0x9d, 0x3a, 0xd9, 0x89, 0x5a, 0xda, 0x7c,
	0xbf, 0x7d, 0x01, 0xf9, 0x94, 0x3a, 0x3e, 0x13, 0x7f, 0x1e, 0x5a, 0xc2, 0x6b, 0x28, 0x52, 0x3e,
	0x94, 0x66, 0xd6, 0xb5, 0xba, 0x30, 0xf3, 0xd0, 0x7a, 0x85, 0x57, 0xb0, 0xd9, 0xfb, 0xbe, 0xa7,
	0x36, 0x62, 0x69, 0x52, 0x9c, 0x6a, 0x3c, 0xf4, 0xaa, 0x16, 0x37, 0xe2, 0xf5, 0x34, 0x6f, 0xe2,
	0xee, 0x27, 0x00, 0x00, 0xff, 0xff, 0x3d, 0x41, 0x6a, 0x1c, 0x40, 0x01, 0x00, 0x00,
}
