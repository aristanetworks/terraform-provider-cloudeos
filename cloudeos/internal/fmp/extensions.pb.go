// Copyright (c) 2020 Arista Networks, Inc.
// Use of this source code is governed by the Mozilla Public License Version 2.0
// that can be found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.12.1
// source: fmp/extensions.proto

package fmp

import (
	reflect "reflect"

	proto "github.com/golang/protobuf/proto"
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

var file_fmp_extensions_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptor.MessageOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         51423,
		Name:          "fmp.model",
		Tag:           "bytes,51423,opt,name=model",
		Filename:      "fmp/extensions.proto",
	},
	{
		ExtendedType:  (*descriptor.MessageOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         51424,
		Name:          "fmp.model_key",
		Tag:           "varint,51424,opt,name=model_key",
		Filename:      "fmp/extensions.proto",
	},
	{
		ExtendedType:  (*descriptor.MessageOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         51425,
		Name:          "fmp.custom_filter",
		Tag:           "bytes,51425,opt,name=custom_filter",
		Filename:      "fmp/extensions.proto",
	},
	{
		ExtendedType:  (*descriptor.MessageOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         51426,
		Name:          "fmp.no_default_filter",
		Tag:           "varint,51426,opt,name=no_default_filter",
		Filename:      "fmp/extensions.proto",
	},
	{
		ExtendedType:  (*descriptor.MessageOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         51427,
		Name:          "fmp.require_set_key",
		Tag:           "varint,51427,opt,name=require_set_key",
		Filename:      "fmp/extensions.proto",
	},
	{
		ExtendedType:  (*descriptor.MessageOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         51428,
		Name:          "fmp.unkeyed_model",
		Tag:           "bytes,51428,opt,name=unkeyed_model",
		Filename:      "fmp/extensions.proto",
	},
}

// Extension fields to descriptor.MessageOptions.
var (
	// TODO: will need an official number from Google, just like gNMI extensions
	//       this works for now, though.
	//
	// optional string model = 51423;
	E_Model = &file_fmp_extensions_proto_extTypes[0]
	// optional bool model_key = 51424;
	E_ModelKey = &file_fmp_extensions_proto_extTypes[1]
	// optional string custom_filter = 51425;
	E_CustomFilter = &file_fmp_extensions_proto_extTypes[2]
	// optional bool no_default_filter = 51426;
	E_NoDefaultFilter = &file_fmp_extensions_proto_extTypes[3]
	// optional bool require_set_key = 51427;
	E_RequireSetKey = &file_fmp_extensions_proto_extTypes[4]
	// optional string unkeyed_model = 51428;
	E_UnkeyedModel = &file_fmp_extensions_proto_extTypes[5]
)

var File_fmp_extensions_proto protoreflect.FileDescriptor

var file_fmp_extensions_proto_rawDesc = []byte{
	0x0a, 0x14, 0x66, 0x6d, 0x70, 0x2f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x66, 0x6d, 0x70, 0x1a, 0x20, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x3a, 0x37, 0x0a,
	0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xdf, 0x91, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x3a, 0x3e, 0x0a, 0x09, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x5f,
	0x6b, 0x65, 0x79, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0xe0, 0x91, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x4b, 0x65, 0x79, 0x3a, 0x46, 0x0a, 0x0d, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d,
	0x5f, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xe1, 0x91, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0c, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x3a, 0x4d,
	0x0a, 0x11, 0x6e, 0x6f, 0x5f, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x5f, 0x66, 0x69, 0x6c,
	0x74, 0x65, 0x72, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0xe2, 0x91, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0f, 0x6e, 0x6f,
	0x44, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x3a, 0x49, 0x0a,
	0x0f, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x5f, 0x73, 0x65, 0x74, 0x5f, 0x6b, 0x65, 0x79,
	0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x18, 0xe3, 0x91, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0d, 0x72, 0x65, 0x71, 0x75, 0x69,
	0x72, 0x65, 0x53, 0x65, 0x74, 0x4b, 0x65, 0x79, 0x3a, 0x46, 0x0a, 0x0d, 0x75, 0x6e, 0x6b, 0x65,
	0x79, 0x65, 0x64, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xe4, 0x91, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0c, 0x75, 0x6e, 0x6b, 0x65, 0x79, 0x65, 0x64, 0x4d, 0x6f, 0x64, 0x65, 0x6c,
	0x42, 0x16, 0x5a, 0x14, 0x61, 0x72, 0x69, 0x73, 0x74, 0x61, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x73, 0x2f, 0x66, 0x6d, 0x70, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_fmp_extensions_proto_goTypes = []interface{}{
	(*descriptor.MessageOptions)(nil), // 0: google.protobuf.MessageOptions
}
var file_fmp_extensions_proto_depIdxs = []int32{
	0, // 0: fmp.model:extendee -> google.protobuf.MessageOptions
	0, // 1: fmp.model_key:extendee -> google.protobuf.MessageOptions
	0, // 2: fmp.custom_filter:extendee -> google.protobuf.MessageOptions
	0, // 3: fmp.no_default_filter:extendee -> google.protobuf.MessageOptions
	0, // 4: fmp.require_set_key:extendee -> google.protobuf.MessageOptions
	0, // 5: fmp.unkeyed_model:extendee -> google.protobuf.MessageOptions
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	0, // [0:6] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_fmp_extensions_proto_init() }
func file_fmp_extensions_proto_init() {
	if File_fmp_extensions_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_fmp_extensions_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 6,
			NumServices:   0,
		},
		GoTypes:           file_fmp_extensions_proto_goTypes,
		DependencyIndexes: file_fmp_extensions_proto_depIdxs,
		ExtensionInfos:    file_fmp_extensions_proto_extTypes,
	}.Build()
	File_fmp_extensions_proto = out.File
	file_fmp_extensions_proto_rawDesc = nil
	file_fmp_extensions_proto_goTypes = nil
	file_fmp_extensions_proto_depIdxs = nil
}
