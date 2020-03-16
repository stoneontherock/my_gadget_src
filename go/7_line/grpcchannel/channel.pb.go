// Code generated by protoc-gen-go. DO NOT EDIT.
// source: channel.proto

package grpcchannel

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// shared empty response
type EmptyResp struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *EmptyResp) Reset()         { *m = EmptyResp{} }
func (m *EmptyResp) String() string { return proto.CompactTextString(m) }
func (*EmptyResp) ProtoMessage()    {}
func (*EmptyResp) Descriptor() ([]byte, []int) {
	return fileDescriptor_c8f385724121f37b, []int{0}
}

func (m *EmptyResp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EmptyResp.Unmarshal(m, b)
}
func (m *EmptyResp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EmptyResp.Marshal(b, m, deterministic)
}
func (m *EmptyResp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EmptyResp.Merge(m, src)
}
func (m *EmptyResp) XXX_Size() int {
	return xxx_messageInfo_EmptyResp.Size(m)
}
func (m *EmptyResp) XXX_DiscardUnknown() {
	xxx_messageInfo_EmptyResp.DiscardUnknown(m)
}

var xxx_messageInfo_EmptyResp proto.InternalMessageInfo

//Ping IN
type Ping struct {
	Mid                  string   `protobuf:"bytes,1,opt,name=mid,proto3" json:"mid,omitempty"`
	Kernel               string   `protobuf:"bytes,2,opt,name=kernel,proto3" json:"kernel,omitempty"`
	OsInfo               string   `protobuf:"bytes,3,opt,name=os_info,json=osInfo,proto3" json:"os_info,omitempty"`
	Interval             int32    `protobuf:"varint,4,opt,name=interval,proto3" json:"interval,omitempty"`
	StartAt              int32    `protobuf:"varint,5,opt,name=start_at,json=startAt,proto3" json:"start_at,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Ping) Reset()         { *m = Ping{} }
func (m *Ping) String() string { return proto.CompactTextString(m) }
func (*Ping) ProtoMessage()    {}
func (*Ping) Descriptor() ([]byte, []int) {
	return fileDescriptor_c8f385724121f37b, []int{1}
}

func (m *Ping) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Ping.Unmarshal(m, b)
}
func (m *Ping) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Ping.Marshal(b, m, deterministic)
}
func (m *Ping) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Ping.Merge(m, src)
}
func (m *Ping) XXX_Size() int {
	return xxx_messageInfo_Ping.Size(m)
}
func (m *Ping) XXX_DiscardUnknown() {
	xxx_messageInfo_Ping.DiscardUnknown(m)
}

var xxx_messageInfo_Ping proto.InternalMessageInfo

func (m *Ping) GetMid() string {
	if m != nil {
		return m.Mid
	}
	return ""
}

func (m *Ping) GetKernel() string {
	if m != nil {
		return m.Kernel
	}
	return ""
}

func (m *Ping) GetOsInfo() string {
	if m != nil {
		return m.OsInfo
	}
	return ""
}

func (m *Ping) GetInterval() int32 {
	if m != nil {
		return m.Interval
	}
	return 0
}

func (m *Ping) GetStartAt() int32 {
	if m != nil {
		return m.StartAt
	}
	return 0
}

//ping out
type Pong struct {
	Action               string   `protobuf:"bytes,1,opt,name=action,proto3" json:"action,omitempty"`
	Data                 []byte   `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Pong) Reset()         { *m = Pong{} }
func (m *Pong) String() string { return proto.CompactTextString(m) }
func (*Pong) ProtoMessage()    {}
func (*Pong) Descriptor() ([]byte, []int) {
	return fileDescriptor_c8f385724121f37b, []int{2}
}

func (m *Pong) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Pong.Unmarshal(m, b)
}
func (m *Pong) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Pong.Marshal(b, m, deterministic)
}
func (m *Pong) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Pong.Merge(m, src)
}
func (m *Pong) XXX_Size() int {
	return xxx_messageInfo_Pong.Size(m)
}
func (m *Pong) XXX_DiscardUnknown() {
	xxx_messageInfo_Pong.DiscardUnknown(m)
}

var xxx_messageInfo_Pong proto.InternalMessageInfo

func (m *Pong) GetAction() string {
	if m != nil {
		return m.Action
	}
	return ""
}

func (m *Pong) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

//cmd
type CmdOutput struct {
	ReturnCode           int32    `protobuf:"zigzag32,1,opt,name=return_code,json=returnCode,proto3" json:"return_code,omitempty"`
	Stdout               string   `protobuf:"bytes,2,opt,name=stdout,proto3" json:"stdout,omitempty"`
	Stderr               string   `protobuf:"bytes,3,opt,name=stderr,proto3" json:"stderr,omitempty"`
	Mid                  string   `protobuf:"bytes,4,opt,name=mid,proto3" json:"mid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CmdOutput) Reset()         { *m = CmdOutput{} }
func (m *CmdOutput) String() string { return proto.CompactTextString(m) }
func (*CmdOutput) ProtoMessage()    {}
func (*CmdOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_c8f385724121f37b, []int{3}
}

func (m *CmdOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CmdOutput.Unmarshal(m, b)
}
func (m *CmdOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CmdOutput.Marshal(b, m, deterministic)
}
func (m *CmdOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CmdOutput.Merge(m, src)
}
func (m *CmdOutput) XXX_Size() int {
	return xxx_messageInfo_CmdOutput.Size(m)
}
func (m *CmdOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_CmdOutput.DiscardUnknown(m)
}

var xxx_messageInfo_CmdOutput proto.InternalMessageInfo

func (m *CmdOutput) GetReturnCode() int32 {
	if m != nil {
		return m.ReturnCode
	}
	return 0
}

func (m *CmdOutput) GetStdout() string {
	if m != nil {
		return m.Stdout
	}
	return ""
}

func (m *CmdOutput) GetStderr() string {
	if m != nil {
		return m.Stderr
	}
	return ""
}

func (m *CmdOutput) GetMid() string {
	if m != nil {
		return m.Mid
	}
	return ""
}

// rpxy Req
type RPxyReq struct {
	Mid                  string   `protobuf:"bytes,1,opt,name=mid,proto3" json:"mid,omitempty"`
	Port2                string   `protobuf:"bytes,2,opt,name=port2,proto3" json:"port2,omitempty"`
	Addr3                string   `protobuf:"bytes,3,opt,name=addr3,proto3" json:"addr3,omitempty"`
	NumOfConn2           int32    `protobuf:"varint,4,opt,name=num_of_conn2,json=numOfConn2,proto3" json:"num_of_conn2,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RPxyReq) Reset()         { *m = RPxyReq{} }
func (m *RPxyReq) String() string { return proto.CompactTextString(m) }
func (*RPxyReq) ProtoMessage()    {}
func (*RPxyReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_c8f385724121f37b, []int{4}
}

func (m *RPxyReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RPxyReq.Unmarshal(m, b)
}
func (m *RPxyReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RPxyReq.Marshal(b, m, deterministic)
}
func (m *RPxyReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RPxyReq.Merge(m, src)
}
func (m *RPxyReq) XXX_Size() int {
	return xxx_messageInfo_RPxyReq.Size(m)
}
func (m *RPxyReq) XXX_DiscardUnknown() {
	xxx_messageInfo_RPxyReq.DiscardUnknown(m)
}

var xxx_messageInfo_RPxyReq proto.InternalMessageInfo

func (m *RPxyReq) GetMid() string {
	if m != nil {
		return m.Mid
	}
	return ""
}

func (m *RPxyReq) GetPort2() string {
	if m != nil {
		return m.Port2
	}
	return ""
}

func (m *RPxyReq) GetAddr3() string {
	if m != nil {
		return m.Addr3
	}
	return ""
}

func (m *RPxyReq) GetNumOfConn2() int32 {
	if m != nil {
		return m.NumOfConn2
	}
	return 0
}

// rpxy resp
type RPxyResp struct {
	Port2                string   `protobuf:"bytes,1,opt,name=port2,proto3" json:"port2,omitempty"`
	Addr3                string   `protobuf:"bytes,2,opt,name=addr3,proto3" json:"addr3,omitempty"`
	NumOfConn2           int32    `protobuf:"varint,3,opt,name=num_of_conn2,json=numOfConn2,proto3" json:"num_of_conn2,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RPxyResp) Reset()         { *m = RPxyResp{} }
func (m *RPxyResp) String() string { return proto.CompactTextString(m) }
func (*RPxyResp) ProtoMessage()    {}
func (*RPxyResp) Descriptor() ([]byte, []int) {
	return fileDescriptor_c8f385724121f37b, []int{5}
}

func (m *RPxyResp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RPxyResp.Unmarshal(m, b)
}
func (m *RPxyResp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RPxyResp.Marshal(b, m, deterministic)
}
func (m *RPxyResp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RPxyResp.Merge(m, src)
}
func (m *RPxyResp) XXX_Size() int {
	return xxx_messageInfo_RPxyResp.Size(m)
}
func (m *RPxyResp) XXX_DiscardUnknown() {
	xxx_messageInfo_RPxyResp.DiscardUnknown(m)
}

var xxx_messageInfo_RPxyResp proto.InternalMessageInfo

func (m *RPxyResp) GetPort2() string {
	if m != nil {
		return m.Port2
	}
	return ""
}

func (m *RPxyResp) GetAddr3() string {
	if m != nil {
		return m.Addr3
	}
	return ""
}

func (m *RPxyResp) GetNumOfConn2() int32 {
	if m != nil {
		return m.NumOfConn2
	}
	return 0
}

func init() {
	proto.RegisterType((*EmptyResp)(nil), "grpcchannel.EmptyResp")
	proto.RegisterType((*Ping)(nil), "grpcchannel.Ping")
	proto.RegisterType((*Pong)(nil), "grpcchannel.Pong")
	proto.RegisterType((*CmdOutput)(nil), "grpcchannel.CmdOutput")
	proto.RegisterType((*RPxyReq)(nil), "grpcchannel.RPxyReq")
	proto.RegisterType((*RPxyResp)(nil), "grpcchannel.RPxyResp")
}

func init() { proto.RegisterFile("channel.proto", fileDescriptor_c8f385724121f37b) }

var fileDescriptor_c8f385724121f37b = []byte{
	// 404 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x92, 0x3f, 0x6f, 0xdb, 0x30,
	0x10, 0xc5, 0xa3, 0x58, 0xfe, 0x77, 0x4e, 0x81, 0xe4, 0x90, 0xa6, 0xaa, 0x97, 0x1a, 0x9a, 0x32,
	0x19, 0x81, 0x32, 0x77, 0x28, 0x84, 0x0e, 0x9d, 0x62, 0x70, 0xea, 0x26, 0xa8, 0x12, 0xed, 0x0a,
	0x95, 0xee, 0x54, 0xea, 0x54, 0xc4, 0x5b, 0xbf, 0x5a, 0xbf, 0x59, 0x21, 0x92, 0x76, 0xea, 0xc0,
	0xe8, 0xc6, 0xf7, 0x48, 0xfd, 0xc8, 0x7b, 0x7a, 0xf0, 0xa6, 0xf8, 0x9e, 0x13, 0xe9, 0x7a, 0xdd,
	0x1a, 0x16, 0xc6, 0xc5, 0xce, 0xb4, 0x85, 0xb7, 0xe2, 0x05, 0xcc, 0x3f, 0x37, 0xad, 0xec, 0x95,
	0xee, 0xda, 0xf8, 0x77, 0x00, 0xe1, 0xa6, 0xa2, 0x1d, 0x5e, 0xc3, 0xa8, 0xa9, 0xca, 0x28, 0x58,
	0x05, 0xf7, 0x73, 0x35, 0x2c, 0xf1, 0x0e, 0x26, 0x3f, 0xb4, 0x21, 0x5d, 0x47, 0x97, 0xd6, 0xf4,
	0x0a, 0xdf, 0xc1, 0x94, 0xbb, 0xac, 0xa2, 0x2d, 0x47, 0x23, 0xb7, 0xc1, 0xdd, 0x17, 0xda, 0x32,
	0x2e, 0x61, 0x56, 0x91, 0x68, 0xf3, 0x2b, 0xaf, 0xa3, 0x70, 0x15, 0xdc, 0x8f, 0xd5, 0x51, 0xe3,
	0x7b, 0x98, 0x75, 0x92, 0x1b, 0xc9, 0x72, 0x89, 0xc6, 0x76, 0x6f, 0x6a, 0xf5, 0x27, 0x89, 0x13,
	0x08, 0x37, 0x4c, 0xbb, 0xe1, 0xbe, 0xbc, 0x90, 0x8a, 0xc9, 0x3f, 0xc2, 0x2b, 0x44, 0x08, 0xcb,
	0x5c, 0x72, 0xfb, 0x8a, 0x2b, 0x65, 0xd7, 0x31, 0xc1, 0x3c, 0x6d, 0xca, 0xa7, 0x5e, 0xda, 0x5e,
	0xf0, 0x03, 0x2c, 0x8c, 0x96, 0xde, 0x50, 0x56, 0x70, 0xa9, 0xed, 0xd7, 0x37, 0x0a, 0x9c, 0x95,
	0x72, 0xa9, 0x07, 0x72, 0x27, 0x25, 0xf7, 0x72, 0x98, 0xc4, 0x29, 0xef, 0x6b, 0x63, 0x0e, 0x83,
	0x38, 0x75, 0xc8, 0x22, 0x3c, 0x66, 0x11, 0x57, 0x30, 0x55, 0x9b, 0xe7, 0xbd, 0xd2, 0x3f, 0xcf,
	0x04, 0x75, 0x0b, 0xe3, 0x96, 0x8d, 0x24, 0x9e, 0xee, 0xc4, 0xe0, 0xe6, 0x65, 0x69, 0x1e, 0x3d,
	0xdb, 0x09, 0x5c, 0xc1, 0x15, 0xf5, 0x4d, 0xc6, 0xdb, 0xac, 0x60, 0xa2, 0xc4, 0xe7, 0x04, 0xd4,
	0x37, 0x4f, 0xdb, 0x74, 0x70, 0xe2, 0xaf, 0x30, 0x73, 0x57, 0x75, 0xed, 0x0b, 0x39, 0x38, 0x4b,
	0xbe, 0xfc, 0x1f, 0x79, 0xf4, 0x9a, 0x9c, 0xfc, 0x09, 0x60, 0x9a, 0xba, 0x12, 0x60, 0x02, 0x13,
	0xa5, 0x07, 0x1c, 0xde, 0xac, 0xff, 0x29, 0xc7, 0x7a, 0xe8, 0xc2, 0xf2, 0x95, 0xc5, 0xb4, 0x8b,
	0x2f, 0x1e, 0x02, 0xfc, 0x68, 0x43, 0x57, 0xba, 0xeb, 0x6b, 0xc1, 0xbb, 0x93, 0x33, 0xc7, 0x9f,
	0xb1, 0x3c, 0xf5, 0x5f, 0x8a, 0x76, 0x81, 0x29, 0x5c, 0xab, 0x8d, 0xe1, 0xe7, 0x7d, 0xca, 0x24,
	0x86, 0xeb, 0x5a, 0x1b, 0xbc, 0x3d, 0x39, 0xed, 0x23, 0x5e, 0xbe, 0x3d, 0xe3, 0x0e, 0x88, 0x87,
	0xe0, 0xdb, 0xc4, 0x16, 0xfa, 0xf1, 0x6f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x70, 0xfe, 0xa4, 0xa9,
	0xe1, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ChannelClient is the client API for Channel service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ChannelClient interface {
	Report(ctx context.Context, in *Ping, opts ...grpc.CallOption) (Channel_ReportClient, error)
	CmdResult(ctx context.Context, in *CmdOutput, opts ...grpc.CallOption) (*EmptyResp, error)
	RProxyController(ctx context.Context, in *RPxyReq, opts ...grpc.CallOption) (Channel_RProxyControllerClient, error)
}

type channelClient struct {
	cc *grpc.ClientConn
}

func NewChannelClient(cc *grpc.ClientConn) ChannelClient {
	return &channelClient{cc}
}

func (c *channelClient) Report(ctx context.Context, in *Ping, opts ...grpc.CallOption) (Channel_ReportClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Channel_serviceDesc.Streams[0], "/grpcchannel.Channel/Report", opts...)
	if err != nil {
		return nil, err
	}
	x := &channelReportClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Channel_ReportClient interface {
	Recv() (*Pong, error)
	grpc.ClientStream
}

type channelReportClient struct {
	grpc.ClientStream
}

func (x *channelReportClient) Recv() (*Pong, error) {
	m := new(Pong)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *channelClient) CmdResult(ctx context.Context, in *CmdOutput, opts ...grpc.CallOption) (*EmptyResp, error) {
	out := new(EmptyResp)
	err := c.cc.Invoke(ctx, "/grpcchannel.Channel/CmdResult", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *channelClient) RProxyController(ctx context.Context, in *RPxyReq, opts ...grpc.CallOption) (Channel_RProxyControllerClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Channel_serviceDesc.Streams[1], "/grpcchannel.Channel/RProxyController", opts...)
	if err != nil {
		return nil, err
	}
	x := &channelRProxyControllerClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Channel_RProxyControllerClient interface {
	Recv() (*RPxyResp, error)
	grpc.ClientStream
}

type channelRProxyControllerClient struct {
	grpc.ClientStream
}

func (x *channelRProxyControllerClient) Recv() (*RPxyResp, error) {
	m := new(RPxyResp)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ChannelServer is the server API for Channel service.
type ChannelServer interface {
	Report(*Ping, Channel_ReportServer) error
	CmdResult(context.Context, *CmdOutput) (*EmptyResp, error)
	RProxyController(*RPxyReq, Channel_RProxyControllerServer) error
}

func RegisterChannelServer(s *grpc.Server, srv ChannelServer) {
	s.RegisterService(&_Channel_serviceDesc, srv)
}

func _Channel_Report_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Ping)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ChannelServer).Report(m, &channelReportServer{stream})
}

type Channel_ReportServer interface {
	Send(*Pong) error
	grpc.ServerStream
}

type channelReportServer struct {
	grpc.ServerStream
}

func (x *channelReportServer) Send(m *Pong) error {
	return x.ServerStream.SendMsg(m)
}

func _Channel_CmdResult_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CmdOutput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChannelServer).CmdResult(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcchannel.Channel/CmdResult",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChannelServer).CmdResult(ctx, req.(*CmdOutput))
	}
	return interceptor(ctx, in, info, handler)
}

func _Channel_RProxyController_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(RPxyReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ChannelServer).RProxyController(m, &channelRProxyControllerServer{stream})
}

type Channel_RProxyControllerServer interface {
	Send(*RPxyResp) error
	grpc.ServerStream
}

type channelRProxyControllerServer struct {
	grpc.ServerStream
}

func (x *channelRProxyControllerServer) Send(m *RPxyResp) error {
	return x.ServerStream.SendMsg(m)
}

var _Channel_serviceDesc = grpc.ServiceDesc{
	ServiceName: "grpcchannel.Channel",
	HandlerType: (*ChannelServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CmdResult",
			Handler:    _Channel_CmdResult_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Report",
			Handler:       _Channel_Report_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "RProxyController",
			Handler:       _Channel_RProxyController_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "channel.proto",
}
