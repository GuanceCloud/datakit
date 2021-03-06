// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: trace.proto

package ddtrace

import (
	fmt "fmt"
	math "math"

	proto "github.com/gogo/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = proto.Marshal
	_ = fmt.Errorf
	_ = math.Inf
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type APIDDTrace struct {
	TraceID              uint64    `protobuf:"varint,1,opt,name=traceID,proto3" json:"traceID,omitempty"`
	Spans                []*DDSpan `protobuf:"bytes,2,rep,name=spans,proto3" json:"spans,omitempty"`
	StartTime            int64     `protobuf:"varint,6,opt,name=startTime,proto3" json:"startTime,omitempty"`
	EndTime              int64     `protobuf:"varint,7,opt,name=endTime,proto3" json:"endTime,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *APIDDTrace) Reset()         { *m = APIDDTrace{} }
func (m *APIDDTrace) String() string { return proto.CompactTextString(m) }
func (*APIDDTrace) ProtoMessage()    {}
func (*APIDDTrace) Descriptor() ([]byte, []int) {
	return fileDescriptor_0571941a1d628a80, []int{0}
}

func (m *APIDDTrace) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_APIDDTrace.Unmarshal(m, b)
}

func (m *APIDDTrace) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_APIDDTrace.Marshal(b, m, deterministic)
}

func (m *APIDDTrace) XXX_Merge(src proto.Message) {
	xxx_messageInfo_APIDDTrace.Merge(m, src)
}

func (m *APIDDTrace) XXX_Size() int {
	return xxx_messageInfo_APIDDTrace.Size(m)
}

func (m *APIDDTrace) XXX_DiscardUnknown() {
	xxx_messageInfo_APIDDTrace.DiscardUnknown(m)
}

var xxx_messageInfo_APIDDTrace proto.InternalMessageInfo

func (m *APIDDTrace) GetTraceID() uint64 {
	if m != nil {
		return m.TraceID
	}
	return 0
}

func (m *APIDDTrace) GetSpans() []*DDSpan {
	if m != nil {
		return m.Spans
	}
	return nil
}

func (m *APIDDTrace) GetStartTime() int64 {
	if m != nil {
		return m.StartTime
	}
	return 0
}

func (m *APIDDTrace) GetEndTime() int64 {
	if m != nil {
		return m.EndTime
	}
	return 0
}

func init() {
	proto.RegisterType((*APIDDTrace)(nil), "ddtrace.APIDDTrace")
}

func init() { proto.RegisterFile("trace.proto", fileDescriptor_0571941a1d628a80) }

var fileDescriptor_0571941a1d628a80 = []byte{
	// 216 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x8f, 0xb1, 0x4a, 0x43, 0x31,
	0x14, 0x86, 0x89, 0xd5, 0x5e, 0x4c, 0x07, 0x21, 0x53, 0x10, 0x87, 0xab, 0x20, 0xdc, 0xc5, 0x04,
	0xf4, 0x09, 0xd4, 0x38, 0x74, 0x93, 0x6b, 0x27, 0xb7, 0xd3, 0x24, 0x5c, 0xa2, 0x69, 0x12, 0x92,
	0x93, 0xc1, 0x07, 0xf0, 0xbd, 0xe5, 0xa6, 0x2d, 0x2e, 0x1d, 0xbf, 0xf3, 0xfd, 0xe7, 0x87, 0x9f,
	0xae, 0x30, 0x83, 0xb6, 0x22, 0xe5, 0x88, 0x91, 0x75, 0xc6, 0x34, 0xbc, 0xbe, 0x4d, 0xbe, 0x4e,
	0x2e, 0x14, 0xe9, 0x42, 0xaa, 0x58, 0xe4, 0xe1, 0x2e, 0x4b, 0x82, 0xb0, 0xcf, 0xde, 0xfd, 0x12,
	0x4a, 0x9f, 0xdf, 0xd7, 0x4a, 0x6d, 0x66, 0xc3, 0x38, 0xed, 0x5a, 0x64, 0xad, 0x38, 0xe9, 0xc9,
	0x70, 0x3e, 0x1e, 0x91, 0xdd, 0xd3, 0x8b, 0xf9, 0xad, 0xf0, 0xb3, 0x7e, 0x31, 0xac, 0x1e, 0xaf,
	0xc4, 0xa1, 0x4c, 0x28, 0xf5, 0x91, 0x20, 0x8c, 0x7b, 0xcb, 0x6e, 0xe8, 0x65, 0x41, 0xc8, 0xb8,
	0x71, 0x3b, 0xcb, 0x97, 0x3d, 0x19, 0x16, 0xe3, 0xff, 0x61, 0xae, 0xb7, 0xc1, 0x34, 0xd7, 0x35,
	0x77, 0xc4, 0x97, 0xb7, 0xcf, 0xd7, 0xc9, 0xa1, 0x87, 0xad, 0xf8, 0x72, 0x30, 0xc5, 0xfa, 0x53,
	0x83, 0xd0, 0x71, 0x27, 0xb5, 0x8f, 0xd5, 0x68, 0xc8, 0xf6, 0x01, 0x63, 0xf4, 0x45, 0x1a, 0x40,
	0xf8, 0x76, 0x28, 0x4f, 0x2f, 0xdb, 0x2e, 0xdb, 0xaa, 0xa7, 0xbf, 0x00, 0x00, 0x00, 0xff, 0xff,
	0x11, 0xbb, 0xed, 0xb5, 0x10, 0x01, 0x00, 0x00,
}
