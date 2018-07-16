// Code generated by protoc-gen-go. DO NOT EDIT.
// source: device/device.proto

package device // import "github.com/ESG-USA/Auklet-Client-C/device"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Metrics struct {
	CpuUsage             float64  `protobuf:"fixed64,1,opt,name=cpuUsage,proto3" json:"cpuUsage,omitempty"`
	MemoryUsage          float64  `protobuf:"fixed64,2,opt,name=memoryUsage,proto3" json:"memoryUsage,omitempty"`
	InboundNetwork       uint64   `protobuf:"varint,3,opt,name=inboundNetwork,proto3" json:"inboundNetwork,omitempty"`
	OutboundNetwork      uint64   `protobuf:"varint,4,opt,name=outboundNetwork,proto3" json:"outboundNetwork,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Metrics) Reset()         { *m = Metrics{} }
func (m *Metrics) String() string { return proto.CompactTextString(m) }
func (*Metrics) ProtoMessage()    {}
func (*Metrics) Descriptor() ([]byte, []int) {
	return fileDescriptor_device_357b37533722dd50, []int{0}
}
func (m *Metrics) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Metrics.Unmarshal(m, b)
}
func (m *Metrics) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Metrics.Marshal(b, m, deterministic)
}
func (dst *Metrics) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Metrics.Merge(dst, src)
}
func (m *Metrics) XXX_Size() int {
	return xxx_messageInfo_Metrics.Size(m)
}
func (m *Metrics) XXX_DiscardUnknown() {
	xxx_messageInfo_Metrics.DiscardUnknown(m)
}

var xxx_messageInfo_Metrics proto.InternalMessageInfo

func (m *Metrics) GetCpuUsage() float64 {
	if m != nil {
		return m.CpuUsage
	}
	return 0
}

func (m *Metrics) GetMemoryUsage() float64 {
	if m != nil {
		return m.MemoryUsage
	}
	return 0
}

func (m *Metrics) GetInboundNetwork() uint64 {
	if m != nil {
		return m.InboundNetwork
	}
	return 0
}

func (m *Metrics) GetOutboundNetwork() uint64 {
	if m != nil {
		return m.OutboundNetwork
	}
	return 0
}

func init() {
	proto.RegisterType((*Metrics)(nil), "device.Metrics")
}

func init() { proto.RegisterFile("device/device.proto", fileDescriptor_device_357b37533722dd50) }

var fileDescriptor_device_357b37533722dd50 = []byte{
	// 181 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x4e, 0x49, 0x2d, 0xcb,
	0x4c, 0x4e, 0xd5, 0x87, 0x50, 0x7a, 0x05, 0x45, 0xf9, 0x25, 0xf9, 0x42, 0x6c, 0x10, 0x9e, 0xd2,
	0x4c, 0x46, 0x2e, 0x76, 0xdf, 0xd4, 0x92, 0xa2, 0xcc, 0xe4, 0x62, 0x21, 0x29, 0x2e, 0x8e, 0xe4,
	0x82, 0xd2, 0xd0, 0xe2, 0xc4, 0xf4, 0x54, 0x09, 0x46, 0x05, 0x46, 0x0d, 0xc6, 0x20, 0x38, 0x5f,
	0x48, 0x81, 0x8b, 0x3b, 0x37, 0x35, 0x37, 0xbf, 0xa8, 0x12, 0x22, 0xcd, 0x04, 0x96, 0x46, 0x16,
	0x12, 0x52, 0xe3, 0xe2, 0xcb, 0xcc, 0x4b, 0xca, 0x2f, 0xcd, 0x4b, 0xf1, 0x4b, 0x2d, 0x29, 0xcf,
	0x2f, 0xca, 0x96, 0x60, 0x56, 0x60, 0xd4, 0x60, 0x09, 0x42, 0x13, 0x15, 0xd2, 0xe0, 0xe2, 0xcf,
	0x2f, 0x2d, 0x41, 0x51, 0xc8, 0x02, 0x56, 0x88, 0x2e, 0xec, 0xa4, 0x1d, 0xa5, 0x99, 0x9e, 0x59,
	0x92, 0x51, 0x9a, 0xa4, 0x97, 0x9c, 0x9f, 0xab, 0xef, 0x1a, 0xec, 0xae, 0x1b, 0x1a, 0xec, 0xa8,
	0xef, 0x58, 0x9a, 0x9d, 0x93, 0x5a, 0xa2, 0xeb, 0x9c, 0x93, 0x99, 0x9a, 0x57, 0xa2, 0xeb, 0x0c,
	0xf5, 0x56, 0x12, 0x1b, 0xd8, 0x5f, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x08, 0x8d, 0x87,
	0xbc, 0xee, 0x00, 0x00, 0x00,
}
