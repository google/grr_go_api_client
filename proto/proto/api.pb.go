// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proto/api.proto

/*
Package api is a generated protocol buffer package.

It is generated from these files:
	proto/api.proto

It has these top-level messages:
	VoidArgs
*/
package api

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

type VoidArgs struct {
	XXX_unrecognized []byte `json:"-"`
}

func (m *VoidArgs) Reset()                    { *m = VoidArgs{} }
func (m *VoidArgs) String() string            { return proto.CompactTextString(m) }
func (*VoidArgs) ProtoMessage()               {}
func (*VoidArgs) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func init() {
	proto.RegisterType((*VoidArgs)(nil), "VoidArgs")
}

func init() { proto.RegisterFile("proto/api.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 50 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2f, 0x28, 0xca, 0x2f,
	0xc9, 0xd7, 0x4f, 0x2c, 0xc8, 0xd4, 0x03, 0xb3, 0x94, 0xb8, 0xb8, 0x38, 0xc2, 0xf2, 0x33, 0x53,
	0x1c, 0x8b, 0xd2, 0x8b, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0x59, 0xeb, 0xc0, 0x6e, 0x1d, 0x00,
	0x00, 0x00,
}