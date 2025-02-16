// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.4
// source: internal/storage/request.proto

package storage

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type MapEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   string   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value []string `protobuf:"bytes,2,rep,name=value,proto3" json:"value,omitempty"`
}

func (x *MapEntry) Reset() {
	*x = MapEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_storage_request_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MapEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MapEntry) ProtoMessage() {}

func (x *MapEntry) ProtoReflect() protoreflect.Message {
	mi := &file_internal_storage_request_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MapEntry.ProtoReflect.Descriptor instead.
func (*MapEntry) Descriptor() ([]byte, []int) {
	return file_internal_storage_request_proto_rawDescGZIP(), []int{0}
}

func (x *MapEntry) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *MapEntry) GetValue() []string {
	if x != nil {
		return x.Value
	}
	return nil
}

type Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Method           string      `protobuf:"bytes,1,opt,name=method,proto3" json:"method,omitempty"`
	Url              string      `protobuf:"bytes,2,opt,name=url,proto3" json:"url,omitempty"`
	Proto            string      `protobuf:"bytes,3,opt,name=proto,proto3" json:"proto,omitempty"`
	ProtoMajor       int32       `protobuf:"varint,4,opt,name=proto_major,json=protoMajor,proto3" json:"proto_major,omitempty"`
	ProtoMinor       int32       `protobuf:"varint,5,opt,name=proto_minor,json=protoMinor,proto3" json:"proto_minor,omitempty"`
	Header           []*MapEntry `protobuf:"bytes,6,rep,name=header,proto3" json:"header,omitempty"`
	Body             []byte      `protobuf:"bytes,7,opt,name=body,proto3" json:"body,omitempty"`
	ContentLength    int64       `protobuf:"varint,8,opt,name=content_length,json=contentLength,proto3" json:"content_length,omitempty"`
	TransferEncoding []string    `protobuf:"bytes,9,rep,name=transfer_encoding,json=transferEncoding,proto3" json:"transfer_encoding,omitempty"`
	Close            bool        `protobuf:"varint,10,opt,name=close,proto3" json:"close,omitempty"`
	Host             string      `protobuf:"bytes,11,opt,name=host,proto3" json:"host,omitempty"`
	Form             []*MapEntry `protobuf:"bytes,12,rep,name=form,proto3" json:"form,omitempty"`
	PostForm         []*MapEntry `protobuf:"bytes,13,rep,name=post_form,json=postForm,proto3" json:"post_form,omitempty"`
	RemoteAddr       string      `protobuf:"bytes,14,opt,name=remote_addr,json=remoteAddr,proto3" json:"remote_addr,omitempty"`
	RequestUri       string      `protobuf:"bytes,15,opt,name=request_uri,json=requestUri,proto3" json:"request_uri,omitempty"`
}

func (x *Request) Reset() {
	*x = Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_storage_request_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request) ProtoMessage() {}

func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_internal_storage_request_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request.ProtoReflect.Descriptor instead.
func (*Request) Descriptor() ([]byte, []int) {
	return file_internal_storage_request_proto_rawDescGZIP(), []int{1}
}

func (x *Request) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *Request) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *Request) GetProto() string {
	if x != nil {
		return x.Proto
	}
	return ""
}

func (x *Request) GetProtoMajor() int32 {
	if x != nil {
		return x.ProtoMajor
	}
	return 0
}

func (x *Request) GetProtoMinor() int32 {
	if x != nil {
		return x.ProtoMinor
	}
	return 0
}

func (x *Request) GetHeader() []*MapEntry {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *Request) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

func (x *Request) GetContentLength() int64 {
	if x != nil {
		return x.ContentLength
	}
	return 0
}

func (x *Request) GetTransferEncoding() []string {
	if x != nil {
		return x.TransferEncoding
	}
	return nil
}

func (x *Request) GetClose() bool {
	if x != nil {
		return x.Close
	}
	return false
}

func (x *Request) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *Request) GetForm() []*MapEntry {
	if x != nil {
		return x.Form
	}
	return nil
}

func (x *Request) GetPostForm() []*MapEntry {
	if x != nil {
		return x.PostForm
	}
	return nil
}

func (x *Request) GetRemoteAddr() string {
	if x != nil {
		return x.RemoteAddr
	}
	return ""
}

func (x *Request) GetRequestUri() string {
	if x != nil {
		return x.RequestUri
	}
	return ""
}

var File_internal_storage_request_proto protoreflect.FileDescriptor

var file_internal_storage_request_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61,
	0x67, 0x65, 0x2f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x22, 0x32, 0x0a, 0x08, 0x4d, 0x61, 0x70,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0xe1, 0x03,
	0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x65, 0x74,
	0x68, 0x6f, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f,
	0x64, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x75, 0x72, 0x6c, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1f, 0x0a, 0x0b, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x5f, 0x6d, 0x61, 0x6a, 0x6f, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x4d, 0x61, 0x6a, 0x6f, 0x72, 0x12, 0x1f, 0x0a, 0x0b, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x5f, 0x6d, 0x69, 0x6e, 0x6f, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x0a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x4d, 0x69, 0x6e, 0x6f, 0x72, 0x12, 0x29, 0x0a, 0x06, 0x68,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x73, 0x74,
	0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x4d, 0x61, 0x70, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06,
	0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x07,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x12, 0x25, 0x0a, 0x0e, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x5f, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x0d, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x4c, 0x65, 0x6e, 0x67, 0x74,
	0x68, 0x12, 0x2b, 0x0a, 0x11, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x5f, 0x65, 0x6e,
	0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x18, 0x09, 0x20, 0x03, 0x28, 0x09, 0x52, 0x10, 0x74, 0x72,
	0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x12, 0x14,
	0x0a, 0x05, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x63,
	0x6c, 0x6f, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x18, 0x0b, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x25, 0x0a, 0x04, 0x66, 0x6f, 0x72, 0x6d,
	0x18, 0x0c, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65,
	0x2e, 0x4d, 0x61, 0x70, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x66, 0x6f, 0x72, 0x6d, 0x12,
	0x2e, 0x0a, 0x09, 0x70, 0x6f, 0x73, 0x74, 0x5f, 0x66, 0x6f, 0x72, 0x6d, 0x18, 0x0d, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x11, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x4d, 0x61, 0x70,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x70, 0x6f, 0x73, 0x74, 0x46, 0x6f, 0x72, 0x6d, 0x12,
	0x1f, 0x0a, 0x0b, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x18, 0x0e,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x41, 0x64, 0x64, 0x72,
	0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x75, 0x72, 0x69, 0x18,
	0x0f, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x55, 0x72,
	0x69, 0x42, 0x3f, 0x5a, 0x3d, 0x67, 0x69, 0x74, 0x6c, 0x61, 0x62, 0x2e, 0x6a, 0x69, 0x61, 0x67,
	0x6f, 0x75, 0x79, 0x75, 0x6e, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x63,
	0x61, 0x72, 0x65, 0x2d, 0x74, 0x6f, 0x6f, 0x6c, 0x73, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x6b, 0x69,
	0x74, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61,
	0x67, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_internal_storage_request_proto_rawDescOnce sync.Once
	file_internal_storage_request_proto_rawDescData = file_internal_storage_request_proto_rawDesc
)

func file_internal_storage_request_proto_rawDescGZIP() []byte {
	file_internal_storage_request_proto_rawDescOnce.Do(func() {
		file_internal_storage_request_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_storage_request_proto_rawDescData)
	})
	return file_internal_storage_request_proto_rawDescData
}

var file_internal_storage_request_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_internal_storage_request_proto_goTypes = []interface{}{
	(*MapEntry)(nil), // 0: storage.MapEntry
	(*Request)(nil),  // 1: storage.Request
}
var file_internal_storage_request_proto_depIdxs = []int32{
	0, // 0: storage.Request.header:type_name -> storage.MapEntry
	0, // 1: storage.Request.form:type_name -> storage.MapEntry
	0, // 2: storage.Request.post_form:type_name -> storage.MapEntry
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_internal_storage_request_proto_init() }
func file_internal_storage_request_proto_init() {
	if File_internal_storage_request_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_storage_request_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MapEntry); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_storage_request_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Request); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_internal_storage_request_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_internal_storage_request_proto_goTypes,
		DependencyIndexes: file_internal_storage_request_proto_depIdxs,
		MessageInfos:      file_internal_storage_request_proto_msgTypes,
	}.Build()
	File_internal_storage_request_proto = out.File
	file_internal_storage_request_proto_rawDesc = nil
	file_internal_storage_request_proto_goTypes = nil
	file_internal_storage_request_proto_depIdxs = nil
}
