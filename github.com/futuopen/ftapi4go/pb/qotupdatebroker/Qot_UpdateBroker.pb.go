// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v5.29.0--rc1
// source: Qot_UpdateBroker.proto

package qotupdatebroker

import (
	_ "github.com/futuopen/ftapi4go/pb/common"
	qotcommon "github.com/futuopen/ftapi4go/pb/qotcommon"
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

type S2C struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Security      *qotcommon.Security `protobuf:"bytes,1,req,name=security" json:"security,omitempty"`           //股票
	Name          *string             `protobuf:"bytes,4,opt,name=name" json:"name,omitempty"`                   // 股票名称
	BrokerAskList []*qotcommon.Broker `protobuf:"bytes,2,rep,name=brokerAskList" json:"brokerAskList,omitempty"` //经纪Ask(卖)盘
	BrokerBidList []*qotcommon.Broker `protobuf:"bytes,3,rep,name=brokerBidList" json:"brokerBidList,omitempty"` //经纪Bid(买)盘
}

func (x *S2C) Reset() {
	*x = S2C{}
	mi := &file_Qot_UpdateBroker_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *S2C) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*S2C) ProtoMessage() {}

func (x *S2C) ProtoReflect() protoreflect.Message {
	mi := &file_Qot_UpdateBroker_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use S2C.ProtoReflect.Descriptor instead.
func (*S2C) Descriptor() ([]byte, []int) {
	return file_Qot_UpdateBroker_proto_rawDescGZIP(), []int{0}
}

func (x *S2C) GetSecurity() *qotcommon.Security {
	if x != nil {
		return x.Security
	}
	return nil
}

func (x *S2C) GetName() string {
	if x != nil && x.Name != nil {
		return *x.Name
	}
	return ""
}

func (x *S2C) GetBrokerAskList() []*qotcommon.Broker {
	if x != nil {
		return x.BrokerAskList
	}
	return nil
}

func (x *S2C) GetBrokerBidList() []*qotcommon.Broker {
	if x != nil {
		return x.BrokerBidList
	}
	return nil
}

type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RetType *int32  `protobuf:"varint,1,req,name=retType,def=-400" json:"retType,omitempty"` //RetType,返回结果
	RetMsg  *string `protobuf:"bytes,2,opt,name=retMsg" json:"retMsg,omitempty"`
	ErrCode *int32  `protobuf:"varint,3,opt,name=errCode" json:"errCode,omitempty"`
	S2C     *S2C    `protobuf:"bytes,4,opt,name=s2c" json:"s2c,omitempty"`
}

// Default values for Response fields.
const (
	Default_Response_RetType = int32(-400)
)

func (x *Response) Reset() {
	*x = Response{}
	mi := &file_Qot_UpdateBroker_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_Qot_UpdateBroker_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_Qot_UpdateBroker_proto_rawDescGZIP(), []int{1}
}

func (x *Response) GetRetType() int32 {
	if x != nil && x.RetType != nil {
		return *x.RetType
	}
	return Default_Response_RetType
}

func (x *Response) GetRetMsg() string {
	if x != nil && x.RetMsg != nil {
		return *x.RetMsg
	}
	return ""
}

func (x *Response) GetErrCode() int32 {
	if x != nil && x.ErrCode != nil {
		return *x.ErrCode
	}
	return 0
}

func (x *Response) GetS2C() *S2C {
	if x != nil {
		return x.S2C
	}
	return nil
}

var File_Qot_UpdateBroker_proto protoreflect.FileDescriptor

var file_Qot_UpdateBroker_proto_rawDesc = []byte{
	0x0a, 0x16, 0x51, 0x6f, 0x74, 0x5f, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x42, 0x72, 0x6f, 0x6b,
	0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10, 0x51, 0x6f, 0x74, 0x5f, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x1a, 0x0c, 0x43, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x10, 0x51, 0x6f, 0x74, 0x5f, 0x43, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xbf, 0x01, 0x0a, 0x03, 0x53,
	0x32, 0x43, 0x12, 0x30, 0x0a, 0x08, 0x73, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x18, 0x01,
	0x20, 0x02, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x51, 0x6f, 0x74, 0x5f, 0x43, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2e, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x52, 0x08, 0x73, 0x65, 0x63, 0x75,
	0x72, 0x69, 0x74, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x38, 0x0a, 0x0d, 0x62, 0x72, 0x6f, 0x6b,
	0x65, 0x72, 0x41, 0x73, 0x6b, 0x4c, 0x69, 0x73, 0x74, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x12, 0x2e, 0x51, 0x6f, 0x74, 0x5f, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x42, 0x72, 0x6f,
	0x6b, 0x65, 0x72, 0x52, 0x0d, 0x62, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x41, 0x73, 0x6b, 0x4c, 0x69,
	0x73, 0x74, 0x12, 0x38, 0x0a, 0x0d, 0x62, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x42, 0x69, 0x64, 0x4c,
	0x69, 0x73, 0x74, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x51, 0x6f, 0x74, 0x5f,
	0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x52, 0x0d, 0x62,
	0x72, 0x6f, 0x6b, 0x65, 0x72, 0x42, 0x69, 0x64, 0x4c, 0x69, 0x73, 0x74, 0x22, 0x85, 0x01, 0x0a,
	0x08, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1e, 0x0a, 0x07, 0x72, 0x65, 0x74,
	0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x02, 0x28, 0x05, 0x3a, 0x04, 0x2d, 0x34, 0x30, 0x30,
	0x52, 0x07, 0x72, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x74,
	0x4d, 0x73, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x74, 0x4d, 0x73,
	0x67, 0x12, 0x18, 0x0a, 0x07, 0x65, 0x72, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x07, 0x65, 0x72, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x27, 0x0a, 0x03, 0x73,
	0x32, 0x63, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x51, 0x6f, 0x74, 0x5f, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x2e, 0x53, 0x32, 0x43, 0x52,
	0x03, 0x73, 0x32, 0x63, 0x42, 0x46, 0x0a, 0x13, 0x63, 0x6f, 0x6d, 0x2e, 0x66, 0x75, 0x74, 0x75,
	0x2e, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x62, 0x5a, 0x2f, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x75, 0x74, 0x75, 0x6f, 0x70, 0x65, 0x6e,
	0x2f, 0x66, 0x74, 0x61, 0x70, 0x69, 0x34, 0x67, 0x6f, 0x2f, 0x70, 0x62, 0x2f, 0x71, 0x6f, 0x74,
	0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x62, 0x72, 0x6f, 0x6b, 0x65, 0x72,
}

var (
	file_Qot_UpdateBroker_proto_rawDescOnce sync.Once
	file_Qot_UpdateBroker_proto_rawDescData = file_Qot_UpdateBroker_proto_rawDesc
)

func file_Qot_UpdateBroker_proto_rawDescGZIP() []byte {
	file_Qot_UpdateBroker_proto_rawDescOnce.Do(func() {
		file_Qot_UpdateBroker_proto_rawDescData = protoimpl.X.CompressGZIP(file_Qot_UpdateBroker_proto_rawDescData)
	})
	return file_Qot_UpdateBroker_proto_rawDescData
}

var file_Qot_UpdateBroker_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_Qot_UpdateBroker_proto_goTypes = []any{
	(*S2C)(nil),                // 0: Qot_UpdateBroker.S2C
	(*Response)(nil),           // 1: Qot_UpdateBroker.Response
	(*qotcommon.Security)(nil), // 2: Qot_Common.Security
	(*qotcommon.Broker)(nil),   // 3: Qot_Common.Broker
}
var file_Qot_UpdateBroker_proto_depIdxs = []int32{
	2, // 0: Qot_UpdateBroker.S2C.security:type_name -> Qot_Common.Security
	3, // 1: Qot_UpdateBroker.S2C.brokerAskList:type_name -> Qot_Common.Broker
	3, // 2: Qot_UpdateBroker.S2C.brokerBidList:type_name -> Qot_Common.Broker
	0, // 3: Qot_UpdateBroker.Response.s2c:type_name -> Qot_UpdateBroker.S2C
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_Qot_UpdateBroker_proto_init() }
func file_Qot_UpdateBroker_proto_init() {
	if File_Qot_UpdateBroker_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_Qot_UpdateBroker_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_Qot_UpdateBroker_proto_goTypes,
		DependencyIndexes: file_Qot_UpdateBroker_proto_depIdxs,
		MessageInfos:      file_Qot_UpdateBroker_proto_msgTypes,
	}.Build()
	File_Qot_UpdateBroker_proto = out.File
	file_Qot_UpdateBroker_proto_rawDesc = nil
	file_Qot_UpdateBroker_proto_goTypes = nil
	file_Qot_UpdateBroker_proto_depIdxs = nil
}
