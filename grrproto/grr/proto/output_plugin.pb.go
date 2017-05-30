// Code generated by protoc-gen-go. DO NOT EDIT.
// source: grr/proto/output_plugin.proto

/*
Package output_plugin is a generated protocol buffer package.

It is generated from these files:
	grr/proto/output_plugin.proto

It has these top-level messages:
	OutputPluginDescriptor
	OutputPluginState
	OutputPluginBatchProcessingStatus
	OutputPluginVerificationResult
	OutputPluginVerificationResultsList
	EmailOutputPluginArgs
	BigQueryOutputPluginArgs
*/
package output_plugin

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import export "github.com/google/grr_go_api_client/grrproto/grr/proto/export_proto"
import jobs "github.com/google/grr_go_api_client/grrproto/grr/proto/jobs_proto"
import _ "github.com/google/grr_go_api_client/grrproto/grr/proto/semantic_proto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type OutputPluginBatchProcessingStatus_Status int32

const (
	OutputPluginBatchProcessingStatus_SUCCESS OutputPluginBatchProcessingStatus_Status = 0
	OutputPluginBatchProcessingStatus_ERROR   OutputPluginBatchProcessingStatus_Status = 1
)

var OutputPluginBatchProcessingStatus_Status_name = map[int32]string{
	0: "SUCCESS",
	1: "ERROR",
}
var OutputPluginBatchProcessingStatus_Status_value = map[string]int32{
	"SUCCESS": 0,
	"ERROR":   1,
}

func (x OutputPluginBatchProcessingStatus_Status) Enum() *OutputPluginBatchProcessingStatus_Status {
	p := new(OutputPluginBatchProcessingStatus_Status)
	*p = x
	return p
}
func (x OutputPluginBatchProcessingStatus_Status) String() string {
	return proto.EnumName(OutputPluginBatchProcessingStatus_Status_name, int32(x))
}
func (x *OutputPluginBatchProcessingStatus_Status) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(OutputPluginBatchProcessingStatus_Status_value, data, "OutputPluginBatchProcessingStatus_Status")
	if err != nil {
		return err
	}
	*x = OutputPluginBatchProcessingStatus_Status(value)
	return nil
}
func (OutputPluginBatchProcessingStatus_Status) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor0, []int{2, 0}
}

type OutputPluginVerificationResult_Status int32

const (
	OutputPluginVerificationResult_N_A     OutputPluginVerificationResult_Status = 0
	OutputPluginVerificationResult_SUCCESS OutputPluginVerificationResult_Status = 1
	OutputPluginVerificationResult_WARNING OutputPluginVerificationResult_Status = 2
	OutputPluginVerificationResult_FAILURE OutputPluginVerificationResult_Status = 3
)

var OutputPluginVerificationResult_Status_name = map[int32]string{
	0: "N_A",
	1: "SUCCESS",
	2: "WARNING",
	3: "FAILURE",
}
var OutputPluginVerificationResult_Status_value = map[string]int32{
	"N_A":     0,
	"SUCCESS": 1,
	"WARNING": 2,
	"FAILURE": 3,
}

func (x OutputPluginVerificationResult_Status) Enum() *OutputPluginVerificationResult_Status {
	p := new(OutputPluginVerificationResult_Status)
	*p = x
	return p
}
func (x OutputPluginVerificationResult_Status) String() string {
	return proto.EnumName(OutputPluginVerificationResult_Status_name, int32(x))
}
func (x *OutputPluginVerificationResult_Status) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(OutputPluginVerificationResult_Status_value, data, "OutputPluginVerificationResult_Status")
	if err != nil {
		return err
	}
	*x = OutputPluginVerificationResult_Status(value)
	return nil
}
func (OutputPluginVerificationResult_Status) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor0, []int{3, 0}
}

type OutputPluginDescriptor struct {
	PluginName       *string `protobuf:"bytes,1,opt,name=plugin_name,json=pluginName" json:"plugin_name,omitempty"`
	PluginArgs       []byte  `protobuf:"bytes,2,opt,name=plugin_args,json=pluginArgs" json:"plugin_args,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *OutputPluginDescriptor) Reset()                    { *m = OutputPluginDescriptor{} }
func (m *OutputPluginDescriptor) String() string            { return proto.CompactTextString(m) }
func (*OutputPluginDescriptor) ProtoMessage()               {}
func (*OutputPluginDescriptor) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *OutputPluginDescriptor) GetPluginName() string {
	if m != nil && m.PluginName != nil {
		return *m.PluginName
	}
	return ""
}

func (m *OutputPluginDescriptor) GetPluginArgs() []byte {
	if m != nil {
		return m.PluginArgs
	}
	return nil
}

type OutputPluginState struct {
	PluginDescriptor *OutputPluginDescriptor `protobuf:"bytes,1,opt,name=plugin_descriptor,json=pluginDescriptor" json:"plugin_descriptor,omitempty"`
	PluginState      *jobs.AttributedDict    `protobuf:"bytes,2,opt,name=plugin_state,json=pluginState" json:"plugin_state,omitempty"`
	XXX_unrecognized []byte                  `json:"-"`
}

func (m *OutputPluginState) Reset()                    { *m = OutputPluginState{} }
func (m *OutputPluginState) String() string            { return proto.CompactTextString(m) }
func (*OutputPluginState) ProtoMessage()               {}
func (*OutputPluginState) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *OutputPluginState) GetPluginDescriptor() *OutputPluginDescriptor {
	if m != nil {
		return m.PluginDescriptor
	}
	return nil
}

func (m *OutputPluginState) GetPluginState() *jobs.AttributedDict {
	if m != nil {
		return m.PluginState
	}
	return nil
}

// Next id: 7
type OutputPluginBatchProcessingStatus struct {
	Status           *OutputPluginBatchProcessingStatus_Status `protobuf:"varint,1,opt,name=status,enum=OutputPluginBatchProcessingStatus_Status,def=0" json:"status,omitempty"`
	PluginDescriptor *OutputPluginDescriptor                   `protobuf:"bytes,6,opt,name=plugin_descriptor,json=pluginDescriptor" json:"plugin_descriptor,omitempty"`
	Summary          *string                                   `protobuf:"bytes,3,opt,name=summary" json:"summary,omitempty"`
	BatchIndex       *uint64                                   `protobuf:"varint,4,opt,name=batch_index,json=batchIndex" json:"batch_index,omitempty"`
	BatchSize        *uint64                                   `protobuf:"varint,5,opt,name=batch_size,json=batchSize" json:"batch_size,omitempty"`
	XXX_unrecognized []byte                                    `json:"-"`
}

func (m *OutputPluginBatchProcessingStatus) Reset()         { *m = OutputPluginBatchProcessingStatus{} }
func (m *OutputPluginBatchProcessingStatus) String() string { return proto.CompactTextString(m) }
func (*OutputPluginBatchProcessingStatus) ProtoMessage()    {}
func (*OutputPluginBatchProcessingStatus) Descriptor() ([]byte, []int) {
	return fileDescriptor0, []int{2}
}

const Default_OutputPluginBatchProcessingStatus_Status OutputPluginBatchProcessingStatus_Status = OutputPluginBatchProcessingStatus_SUCCESS

func (m *OutputPluginBatchProcessingStatus) GetStatus() OutputPluginBatchProcessingStatus_Status {
	if m != nil && m.Status != nil {
		return *m.Status
	}
	return Default_OutputPluginBatchProcessingStatus_Status
}

func (m *OutputPluginBatchProcessingStatus) GetPluginDescriptor() *OutputPluginDescriptor {
	if m != nil {
		return m.PluginDescriptor
	}
	return nil
}

func (m *OutputPluginBatchProcessingStatus) GetSummary() string {
	if m != nil && m.Summary != nil {
		return *m.Summary
	}
	return ""
}

func (m *OutputPluginBatchProcessingStatus) GetBatchIndex() uint64 {
	if m != nil && m.BatchIndex != nil {
		return *m.BatchIndex
	}
	return 0
}

func (m *OutputPluginBatchProcessingStatus) GetBatchSize() uint64 {
	if m != nil && m.BatchSize != nil {
		return *m.BatchSize
	}
	return 0
}

type OutputPluginVerificationResult struct {
	PluginId         *string                                `protobuf:"bytes,1,opt,name=plugin_id,json=pluginId" json:"plugin_id,omitempty"`
	PluginDescriptor *OutputPluginDescriptor                `protobuf:"bytes,2,opt,name=plugin_descriptor,json=pluginDescriptor" json:"plugin_descriptor,omitempty"`
	Status           *OutputPluginVerificationResult_Status `protobuf:"varint,3,opt,name=status,enum=OutputPluginVerificationResult_Status" json:"status,omitempty"`
	StatusMessage    *string                                `protobuf:"bytes,4,opt,name=status_message,json=statusMessage" json:"status_message,omitempty"`
	Timestamp        *uint64                                `protobuf:"varint,5,opt,name=timestamp" json:"timestamp,omitempty"`
	XXX_unrecognized []byte                                 `json:"-"`
}

func (m *OutputPluginVerificationResult) Reset()                    { *m = OutputPluginVerificationResult{} }
func (m *OutputPluginVerificationResult) String() string            { return proto.CompactTextString(m) }
func (*OutputPluginVerificationResult) ProtoMessage()               {}
func (*OutputPluginVerificationResult) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *OutputPluginVerificationResult) GetPluginId() string {
	if m != nil && m.PluginId != nil {
		return *m.PluginId
	}
	return ""
}

func (m *OutputPluginVerificationResult) GetPluginDescriptor() *OutputPluginDescriptor {
	if m != nil {
		return m.PluginDescriptor
	}
	return nil
}

func (m *OutputPluginVerificationResult) GetStatus() OutputPluginVerificationResult_Status {
	if m != nil && m.Status != nil {
		return *m.Status
	}
	return OutputPluginVerificationResult_N_A
}

func (m *OutputPluginVerificationResult) GetStatusMessage() string {
	if m != nil && m.StatusMessage != nil {
		return *m.StatusMessage
	}
	return ""
}

func (m *OutputPluginVerificationResult) GetTimestamp() uint64 {
	if m != nil && m.Timestamp != nil {
		return *m.Timestamp
	}
	return 0
}

type OutputPluginVerificationResultsList struct {
	Results          []*OutputPluginVerificationResult `protobuf:"bytes,1,rep,name=results" json:"results,omitempty"`
	XXX_unrecognized []byte                            `json:"-"`
}

func (m *OutputPluginVerificationResultsList) Reset()         { *m = OutputPluginVerificationResultsList{} }
func (m *OutputPluginVerificationResultsList) String() string { return proto.CompactTextString(m) }
func (*OutputPluginVerificationResultsList) ProtoMessage()    {}
func (*OutputPluginVerificationResultsList) Descriptor() ([]byte, []int) {
	return fileDescriptor0, []int{4}
}

func (m *OutputPluginVerificationResultsList) GetResults() []*OutputPluginVerificationResult {
	if m != nil {
		return m.Results
	}
	return nil
}

type EmailOutputPluginArgs struct {
	EmailAddress     *string `protobuf:"bytes,1,opt,name=email_address,json=emailAddress" json:"email_address,omitempty"`
	EmailsLimit      *uint64 `protobuf:"varint,2,opt,name=emails_limit,json=emailsLimit,def=100" json:"emails_limit,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *EmailOutputPluginArgs) Reset()                    { *m = EmailOutputPluginArgs{} }
func (m *EmailOutputPluginArgs) String() string            { return proto.CompactTextString(m) }
func (*EmailOutputPluginArgs) ProtoMessage()               {}
func (*EmailOutputPluginArgs) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

const Default_EmailOutputPluginArgs_EmailsLimit uint64 = 100

func (m *EmailOutputPluginArgs) GetEmailAddress() string {
	if m != nil && m.EmailAddress != nil {
		return *m.EmailAddress
	}
	return ""
}

func (m *EmailOutputPluginArgs) GetEmailsLimit() uint64 {
	if m != nil && m.EmailsLimit != nil {
		return *m.EmailsLimit
	}
	return Default_EmailOutputPluginArgs_EmailsLimit
}

type BigQueryOutputPluginArgs struct {
	ExportOptions    *export.ExportOptions `protobuf:"bytes,2,opt,name=export_options,json=exportOptions" json:"export_options,omitempty"`
	ConvertValues    *bool                 `protobuf:"varint,3,opt,name=convert_values,json=convertValues,def=1" json:"convert_values,omitempty"`
	XXX_unrecognized []byte                `json:"-"`
}

func (m *BigQueryOutputPluginArgs) Reset()                    { *m = BigQueryOutputPluginArgs{} }
func (m *BigQueryOutputPluginArgs) String() string            { return proto.CompactTextString(m) }
func (*BigQueryOutputPluginArgs) ProtoMessage()               {}
func (*BigQueryOutputPluginArgs) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

const Default_BigQueryOutputPluginArgs_ConvertValues bool = true

func (m *BigQueryOutputPluginArgs) GetExportOptions() *export.ExportOptions {
	if m != nil {
		return m.ExportOptions
	}
	return nil
}

func (m *BigQueryOutputPluginArgs) GetConvertValues() bool {
	if m != nil && m.ConvertValues != nil {
		return *m.ConvertValues
	}
	return Default_BigQueryOutputPluginArgs_ConvertValues
}

func init() {
	proto.RegisterType((*OutputPluginDescriptor)(nil), "OutputPluginDescriptor")
	proto.RegisterType((*OutputPluginState)(nil), "OutputPluginState")
	proto.RegisterType((*OutputPluginBatchProcessingStatus)(nil), "OutputPluginBatchProcessingStatus")
	proto.RegisterType((*OutputPluginVerificationResult)(nil), "OutputPluginVerificationResult")
	proto.RegisterType((*OutputPluginVerificationResultsList)(nil), "OutputPluginVerificationResultsList")
	proto.RegisterType((*EmailOutputPluginArgs)(nil), "EmailOutputPluginArgs")
	proto.RegisterType((*BigQueryOutputPluginArgs)(nil), "BigQueryOutputPluginArgs")
	proto.RegisterEnum("OutputPluginBatchProcessingStatus_Status", OutputPluginBatchProcessingStatus_Status_name, OutputPluginBatchProcessingStatus_Status_value)
	proto.RegisterEnum("OutputPluginVerificationResult_Status", OutputPluginVerificationResult_Status_name, OutputPluginVerificationResult_Status_value)
}

func init() { proto.RegisterFile("grr/proto/output_plugin.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 993 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x55, 0xcd, 0x72, 0x1b, 0x45,
	0x10, 0xce, 0x5a, 0x4e, 0x1c, 0x8d, 0x6c, 0x47, 0x19, 0x20, 0x59, 0xa8, 0x22, 0x0c, 0x0a, 0x71,
	0x0c, 0x24, 0x6b, 0xe3, 0x54, 0xf8, 0x31, 0x10, 0x4a, 0xb2, 0x6c, 0x97, 0x53, 0xfe, 0x63, 0xe4,
	0x24, 0xdc, 0x96, 0xd1, 0x6e, 0x4b, 0x1a, 0xb3, 0x7f, 0xcc, 0xcc, 0x86, 0x28, 0xc5, 0x85, 0x2b,
	0x55, 0x1c, 0x78, 0x08, 0x5e, 0x80, 0x67, 0xe0, 0x21, 0x38, 0x43, 0xf1, 0x08, 0xdc, 0x38, 0x50,
	0xf3, 0xa3, 0x9f, 0xb2, 0x55, 0xce, 0x49, 0x33, 0x5f, 0x77, 0x7f, 0xdd, 0xfd, 0x75, 0xef, 0x08,
	0xbd, 0xdd, 0x17, 0x62, 0xad, 0x10, 0xb9, 0xca, 0xd7, 0xf2, 0x52, 0x15, 0xa5, 0x0a, 0x8b, 0xa4,
	0xec, 0xf3, 0x2c, 0x30, 0xd8, 0x5b, 0x37, 0x26, 0x66, 0x78, 0x51, 0xe4, 0x42, 0x39, 0xfc, 0xf5,
	0x09, 0x7e, 0x9a, 0x77, 0xa5, 0x43, 0xfd, 0x09, 0x2a, 0x21, 0x65, 0x99, 0xe2, 0x91, 0xb5, 0x34,
	0xfe, 0xf5, 0xd0, 0x8d, 0x23, 0xc3, 0x7f, 0x6c, 0xe8, 0xdb, 0x20, 0x23, 0xc1, 0x0b, 0x95, 0x0b,
	0xbc, 0x8b, 0x6a, 0x36, 0x65, 0x98, 0xb1, 0x14, 0x7c, 0x8f, 0x78, 0xab, 0xd5, 0xd6, 0xca, 0x5f,
	0xff, 0xfd, 0xfd, 0x87, 0x47, 0xf0, 0xad, 0x93, 0x01, 0x10, 0x8d, 0x93, 0xbc, 0x47, 0xd4, 0x00,
	0x88, 0x2d, 0x92, 0xb8, 0x22, 0x29, 0xb2, 0x87, 0x43, 0x96, 0x02, 0xfe, 0xd5, 0x1b, 0x33, 0x31,
	0xd1, 0x97, 0xfe, 0x1c, 0xf1, 0x56, 0x17, 0x5b, 0x85, 0x61, 0x3a, 0xc5, 0x27, 0x9a, 0xa9, 0x60,
	0x82, 0xa5, 0xa0, 0x40, 0x48, 0xd2, 0xcb, 0x05, 0x51, 0x03, 0x2e, 0x47, 0x54, 0xe4, 0xa0, 0x94,
	0x8a, 0x74, 0x81, 0xb0, 0x8c, 0xf0, 0x4c, 0x2a, 0x96, 0x45, 0xe3, 0xa4, 0xba, 0x80, 0xd8, 0x39,
	0xde, 0x95, 0x44, 0xb3, 0x87, 0x6a, 0x58, 0x40, 0xf0, 0x01, 0xde, 0x05, 0xd7, 0x51, 0x53, 0xf4,
	0xe5, 0x56, 0xc2, 0xa4, 0x1c, 0xd5, 0xa4, 0x81, 0xc6, 0x2f, 0x1e, 0xba, 0x3e, 0xdd, 0x77, 0x47,
	0x31, 0x05, 0xb8, 0x8d, 0xae, 0xbb, 0x42, 0xe3, 0xb1, 0x0e, 0xa6, 0xf1, 0xda, 0xc6, 0xcd, 0x60,
	0xb6, 0x4c, 0xb4, 0x5e, 0x9c, 0x15, 0x6e, 0x03, 0x2d, 0x3a, 0x16, 0xa9, 0x59, 0x4d, 0xbf, 0xb5,
	0x8d, 0x6b, 0x41, 0x53, 0x29, 0xc1, 0xbb, 0xa5, 0x82, 0xb8, 0xcd, 0x23, 0x45, 0x9d, 0x26, 0x26,
	0x73, 0xe3, 0xb7, 0x79, 0xf4, 0xee, 0x74, 0x82, 0x16, 0x53, 0xd1, 0xe0, 0x58, 0xe4, 0x11, 0x48,
	0xc9, 0xb3, 0xbe, 0x76, 0x2a, 0x25, 0xfe, 0xc9, 0x43, 0x57, 0xa4, 0x39, 0x9a, 0xaa, 0x96, 0x37,
	0xde, 0x0f, 0x5e, 0x19, 0x14, 0xd8, 0x9f, 0xcd, 0x85, 0xce, 0x93, 0xad, 0xad, 0xed, 0x4e, 0xa7,
	0xf5, 0xd0, 0x08, 0xbf, 0x86, 0xef, 0x3f, 0x63, 0xd2, 0x2a, 0xdd, 0xd5, 0x71, 0x44, 0x96, 0x91,
	0x0e, 0xec, 0x95, 0x49, 0x32, 0x24, 0x85, 0x65, 0x81, 0x98, 0xe4, 0x82, 0x64, 0xb9, 0xfa, 0x8a,
	0xba, 0xc4, 0xf8, 0x67, 0x6f, 0x96, 0x48, 0x57, 0x2e, 0x14, 0xa9, 0xd5, 0x36, 0x39, 0x1f, 0xe1,
	0x2f, 0x26, 0xd8, 0xcc, 0xc5, 0x21, 0x02, 0x64, 0x91, 0x67, 0x92, 0x77, 0x13, 0x98, 0x6c, 0x82,
	0x00, 0x59, 0x26, 0x2a, 0x98, 0x29, 0xf5, 0x82, 0x2c, 0xd3, 0x94, 0x89, 0xa1, 0x5f, 0x31, 0xfb,
	0xe9, 0x9b, 0x44, 0x18, 0xd7, 0x3b, 0x16, 0x26, 0x29, 0x48, 0xc9, 0xfa, 0x10, 0xd0, 0x91, 0x23,
	0x3e, 0x40, 0x35, 0xd3, 0x70, 0xc8, 0xb3, 0x18, 0x5e, 0xf8, 0xf3, 0xc4, 0x5b, 0x9d, 0x6f, 0xdd,
	0x33, 0x71, 0x2b, 0xf8, 0xbd, 0xc3, 0x32, 0xed, 0xc2, 0xb8, 0x38, 0xab, 0x4c, 0x17, 0x78, 0xd6,
	0x9f, 0x48, 0x12, 0x50, 0x64, 0xf0, 0x3d, 0x1d, 0x8f, 0x1f, 0x23, 0x7b, 0x0b, 0x25, 0x7f, 0x09,
	0xfe, 0x65, 0xc3, 0xf6, 0xa1, 0x61, 0xbb, 0x83, 0x6f, 0x77, 0xf8, 0x4b, 0x38, 0xcb, 0xa5, 0xce,
	0x90, 0x55, 0x8d, 0x41, 0x7b, 0x36, 0x08, 0xba, 0xe2, 0x26, 0x5d, 0x43, 0xa3, 0x79, 0xd5, 0x2f,
	0xe1, 0x2a, 0xba, 0xbc, 0x4d, 0xe9, 0x11, 0xad, 0x7b, 0x8d, 0x7f, 0x2a, 0xe8, 0xd6, 0xb4, 0xc6,
	0x4f, 0x41, 0xf0, 0x1e, 0x8f, 0x98, 0xe2, 0x79, 0x46, 0x8d, 0x4c, 0xf8, 0x1b, 0x54, 0x75, 0xf3,
	0xe1, 0xb1, 0xfb, 0x6a, 0x3f, 0x37, 0xf5, 0x3c, 0xc4, 0x0f, 0x3a, 0x4a, 0xe8, 0xec, 0x3c, 0x86,
	0x4c, 0xf1, 0xde, 0x50, 0x9f, 0xcf, 0xcf, 0xc0, 0xb6, 0x1b, 0x0d, 0x20, 0xfa, 0x4e, 0xd7, 0x77,
	0xd5, 0xc2, 0x7b, 0x31, 0xfe, 0x71, 0xd6, 0xe4, 0xe7, 0x2e, 0x9e, 0xfc, 0xc7, 0x26, 0xf5, 0x3a,
	0x0e, 0xce, 0x4f, 0xde, 0xa5, 0x53, 0x03, 0xa6, 0xee, 0xca, 0xb3, 0x59, 0xcf, 0xcf, 0xfa, 0xd1,
	0x78, 0xf7, 0x2b, 0x66, 0xf7, 0x57, 0x82, 0x8b, 0x85, 0x70, 0x8b, 0x3f, 0x5e, 0xdc, 0x3b, 0x68,
	0xd9, 0x9e, 0x42, 0xb7, 0x13, 0x66, 0xf4, 0x55, 0xba, 0x64, 0xd1, 0x03, 0x0b, 0xe2, 0xc7, 0xa8,
	0xaa, 0x78, 0x0a, 0x52, 0xb1, 0xb4, 0x70, 0xe3, 0x74, 0xcb, 0x81, 0x6a, 0xb4, 0xbd, 0xd3, 0x66,
	0x0a, 0xb4, 0x1d, 0xdf, 0x3c, 0x19, 0x79, 0x8d, 0xfa, 0x31, 0xb5, 0x07, 0x74, 0x12, 0xde, 0xf8,
	0x74, 0x3c, 0xcf, 0x05, 0x54, 0x39, 0x0c, 0x9b, 0xf5, 0x4b, 0xd3, 0x83, 0xf5, 0xf4, 0xe5, 0x59,
	0x93, 0x1e, 0xee, 0x1d, 0xee, 0xd6, 0xe7, 0xf4, 0x65, 0xa7, 0xb9, 0xb7, 0xff, 0x84, 0x6e, 0xd7,
	0x2b, 0x8d, 0x6f, 0xd1, 0xed, 0x8b, 0xbb, 0x93, 0xfb, 0x5c, 0x2a, 0xfc, 0x19, 0x5a, 0xb0, 0x1f,
	0x87, 0x7e, 0x10, 0x2a, 0xab, 0xb5, 0x8d, 0x77, 0x5e, 0x21, 0x0a, 0x1d, 0xf9, 0x37, 0x7e, 0xf7,
	0xd0, 0x1b, 0xdb, 0x29, 0xe3, 0xc9, 0x74, 0x80, 0x7e, 0x1b, 0xf1, 0xf7, 0x68, 0x09, 0xb4, 0x21,
	0x64, 0x71, 0x2c, 0x40, 0x4a, 0xb7, 0x44, 0xfb, 0x46, 0x85, 0x1d, 0x84, 0xdb, 0x79, 0xca, 0x78,
	0x66, 0x62, 0x9b, 0xd6, 0x03, 0xaf, 0xeb, 0x47, 0xdc, 0x04, 0x11, 0x17, 0x64, 0xc6, 0x3a, 0xfa,
	0xfa, 0x24, 0xf9, 0x81, 0x27, 0x89, 0x7e, 0xc4, 0x25, 0x64, 0x8a, 0xa8, 0x3c, 0xa0, 0x8b, 0x30,
	0x1d, 0xbf, 0x82, 0xec, 0x5d, 0x86, 0x09, 0x4f, 0xb9, 0x32, 0x4b, 0x35, 0xbf, 0x59, 0xf9, 0x68,
	0x7d, 0x9d, 0xd6, 0xac, 0x61, 0x5f, 0xe3, 0x8d, 0x3f, 0x3d, 0xe4, 0xb7, 0x78, 0xff, 0xeb, 0x12,
	0xc4, 0xf0, 0x5c, 0xdd, 0xc7, 0x68, 0xd9, 0xfe, 0x17, 0x86, 0x79, 0xa1, 0x3b, 0x96, 0x6e, 0x37,
	0x97, 0x83, 0x6d, 0x03, 0x1f, 0x59, 0xb4, 0xf5, 0xa6, 0x69, 0xe4, 0x35, 0x7c, 0xcd, 0xc2, 0xc4,
	0x79, 0x07, 0xbe, 0x47, 0x97, 0x60, 0xda, 0x13, 0x9f, 0xa2, 0xe5, 0x28, 0xcf, 0x9e, 0x83, 0x50,
	0xe1, 0x73, 0x96, 0x94, 0x60, 0x57, 0xef, 0xea, 0xe6, 0xbc, 0x12, 0x25, 0xb4, 0xbe, 0x34, 0x3c,
	0x9f, 0xe0, 0x07, 0x7b, 0x3d, 0xa2, 0x81, 0x7b, 0xc4, 0x39, 0x13, 0xeb, 0x6c, 0x1e, 0x30, 0x4b,
	0x78, 0xbf, 0x27, 0x38, 0x64, 0x71, 0x32, 0xd4, 0x58, 0xca, 0x54, 0xe0, 0xcf, 0xd1, 0x25, 0xe7,
	0xfd, 0xd4, 0x38, 0xff, 0x1f, 0x00, 0x00, 0xff, 0xff, 0xd6, 0x25, 0x6d, 0xe3, 0xf1, 0x07, 0x00,
	0x00,
}