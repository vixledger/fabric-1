// Code generated by protoc-gen-go. DO NOT EDIT.
// source: messages.proto

package cosipbft

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	any "github.com/golang/protobuf/ptypes/any"
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

// ForwardLinkProto is the message representing a forward link between
// two proposals. It contains both hash and the prepare and commit
// signatures.
type ForwardLinkProto struct {
	From                 []byte   `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty"`
	To                   []byte   `protobuf:"bytes,2,opt,name=to,proto3" json:"to,omitempty"`
	Prepare              *any.Any `protobuf:"bytes,3,opt,name=prepare,proto3" json:"prepare,omitempty"`
	Commit               *any.Any `protobuf:"bytes,4,opt,name=commit,proto3" json:"commit,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ForwardLinkProto) Reset()         { *m = ForwardLinkProto{} }
func (m *ForwardLinkProto) String() string { return proto.CompactTextString(m) }
func (*ForwardLinkProto) ProtoMessage()    {}
func (*ForwardLinkProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_4dc296cbfe5ffcd5, []int{0}
}

func (m *ForwardLinkProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ForwardLinkProto.Unmarshal(m, b)
}
func (m *ForwardLinkProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ForwardLinkProto.Marshal(b, m, deterministic)
}
func (m *ForwardLinkProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ForwardLinkProto.Merge(m, src)
}
func (m *ForwardLinkProto) XXX_Size() int {
	return xxx_messageInfo_ForwardLinkProto.Size(m)
}
func (m *ForwardLinkProto) XXX_DiscardUnknown() {
	xxx_messageInfo_ForwardLinkProto.DiscardUnknown(m)
}

var xxx_messageInfo_ForwardLinkProto proto.InternalMessageInfo

func (m *ForwardLinkProto) GetFrom() []byte {
	if m != nil {
		return m.From
	}
	return nil
}

func (m *ForwardLinkProto) GetTo() []byte {
	if m != nil {
		return m.To
	}
	return nil
}

func (m *ForwardLinkProto) GetPrepare() *any.Any {
	if m != nil {
		return m.Prepare
	}
	return nil
}

func (m *ForwardLinkProto) GetCommit() *any.Any {
	if m != nil {
		return m.Commit
	}
	return nil
}

// ChainProto is the message representing a list of forward links that creates
// a verifiable chain.
type ChainProto struct {
	Links                []*ForwardLinkProto `protobuf:"bytes,1,rep,name=links,proto3" json:"links,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *ChainProto) Reset()         { *m = ChainProto{} }
func (m *ChainProto) String() string { return proto.CompactTextString(m) }
func (*ChainProto) ProtoMessage()    {}
func (*ChainProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_4dc296cbfe5ffcd5, []int{1}
}

func (m *ChainProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChainProto.Unmarshal(m, b)
}
func (m *ChainProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChainProto.Marshal(b, m, deterministic)
}
func (m *ChainProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChainProto.Merge(m, src)
}
func (m *ChainProto) XXX_Size() int {
	return xxx_messageInfo_ChainProto.Size(m)
}
func (m *ChainProto) XXX_DiscardUnknown() {
	xxx_messageInfo_ChainProto.DiscardUnknown(m)
}

var xxx_messageInfo_ChainProto proto.InternalMessageInfo

func (m *ChainProto) GetLinks() []*ForwardLinkProto {
	if m != nil {
		return m.Links
	}
	return nil
}

// PrepareRequest is the message sent to start a consensus for a proposal.
type PrepareRequest struct {
	Proposal             *any.Any `protobuf:"bytes,1,opt,name=proposal,proto3" json:"proposal,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PrepareRequest) Reset()         { *m = PrepareRequest{} }
func (m *PrepareRequest) String() string { return proto.CompactTextString(m) }
func (*PrepareRequest) ProtoMessage()    {}
func (*PrepareRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4dc296cbfe5ffcd5, []int{2}
}

func (m *PrepareRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PrepareRequest.Unmarshal(m, b)
}
func (m *PrepareRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PrepareRequest.Marshal(b, m, deterministic)
}
func (m *PrepareRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PrepareRequest.Merge(m, src)
}
func (m *PrepareRequest) XXX_Size() int {
	return xxx_messageInfo_PrepareRequest.Size(m)
}
func (m *PrepareRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PrepareRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PrepareRequest proto.InternalMessageInfo

func (m *PrepareRequest) GetProposal() *any.Any {
	if m != nil {
		return m.Proposal
	}
	return nil
}

// CommitRequest is the message sent to commit to a proposal.
type CommitRequest struct {
	To                   []byte   `protobuf:"bytes,1,opt,name=to,proto3" json:"to,omitempty"`
	Prepare              *any.Any `protobuf:"bytes,2,opt,name=prepare,proto3" json:"prepare,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CommitRequest) Reset()         { *m = CommitRequest{} }
func (m *CommitRequest) String() string { return proto.CompactTextString(m) }
func (*CommitRequest) ProtoMessage()    {}
func (*CommitRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4dc296cbfe5ffcd5, []int{3}
}

func (m *CommitRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CommitRequest.Unmarshal(m, b)
}
func (m *CommitRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CommitRequest.Marshal(b, m, deterministic)
}
func (m *CommitRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CommitRequest.Merge(m, src)
}
func (m *CommitRequest) XXX_Size() int {
	return xxx_messageInfo_CommitRequest.Size(m)
}
func (m *CommitRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CommitRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CommitRequest proto.InternalMessageInfo

func (m *CommitRequest) GetTo() []byte {
	if m != nil {
		return m.To
	}
	return nil
}

func (m *CommitRequest) GetPrepare() *any.Any {
	if m != nil {
		return m.Prepare
	}
	return nil
}

// PropagateRequest is the last message of a consensus process to send the valid
// forward link to participants.
type PropagateRequest struct {
	To                   []byte   `protobuf:"bytes,1,opt,name=to,proto3" json:"to,omitempty"`
	Commit               *any.Any `protobuf:"bytes,2,opt,name=commit,proto3" json:"commit,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PropagateRequest) Reset()         { *m = PropagateRequest{} }
func (m *PropagateRequest) String() string { return proto.CompactTextString(m) }
func (*PropagateRequest) ProtoMessage()    {}
func (*PropagateRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4dc296cbfe5ffcd5, []int{4}
}

func (m *PropagateRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PropagateRequest.Unmarshal(m, b)
}
func (m *PropagateRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PropagateRequest.Marshal(b, m, deterministic)
}
func (m *PropagateRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PropagateRequest.Merge(m, src)
}
func (m *PropagateRequest) XXX_Size() int {
	return xxx_messageInfo_PropagateRequest.Size(m)
}
func (m *PropagateRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PropagateRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PropagateRequest proto.InternalMessageInfo

func (m *PropagateRequest) GetTo() []byte {
	if m != nil {
		return m.To
	}
	return nil
}

func (m *PropagateRequest) GetCommit() *any.Any {
	if m != nil {
		return m.Commit
	}
	return nil
}

func init() {
	proto.RegisterType((*ForwardLinkProto)(nil), "cosipbft.ForwardLinkProto")
	proto.RegisterType((*ChainProto)(nil), "cosipbft.ChainProto")
	proto.RegisterType((*PrepareRequest)(nil), "cosipbft.PrepareRequest")
	proto.RegisterType((*CommitRequest)(nil), "cosipbft.CommitRequest")
	proto.RegisterType((*PropagateRequest)(nil), "cosipbft.PropagateRequest")
}

func init() {
	proto.RegisterFile("messages.proto", fileDescriptor_4dc296cbfe5ffcd5)
}

var fileDescriptor_4dc296cbfe5ffcd5 = []byte{
	// 278 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x90, 0x41, 0x4b, 0xc3, 0x30,
	0x14, 0xc7, 0x49, 0x37, 0xe7, 0x78, 0x9b, 0x65, 0x04, 0x0f, 0x75, 0xa7, 0xd2, 0x53, 0x0f, 0x92,
	0x8d, 0x79, 0x14, 0x04, 0x37, 0xf0, 0x24, 0x58, 0x7a, 0xf4, 0x96, 0xce, 0xb4, 0x96, 0xb5, 0x79,
	0x31, 0xc9, 0x90, 0x7d, 0x0f, 0x3f, 0xb0, 0x90, 0x18, 0x05, 0x71, 0xea, 0x2d, 0x3c, 0x7e, 0x79,
	0xff, 0xff, 0xfb, 0x41, 0xdc, 0x0b, 0x63, 0x78, 0x23, 0x0c, 0x53, 0x1a, 0x2d, 0xd2, 0xf1, 0x16,
	0x4d, 0xab, 0xaa, 0xda, 0xce, 0x2f, 0x1a, 0xc4, 0xa6, 0x13, 0x0b, 0x37, 0xaf, 0xf6, 0xf5, 0x82,
	0xcb, 0x83, 0x87, 0xb2, 0x37, 0x02, 0xb3, 0x3b, 0xd4, 0xaf, 0x5c, 0x3f, 0xdd, 0xb7, 0x72, 0x57,
	0xb8, 0x9f, 0x14, 0x86, 0xb5, 0xc6, 0x3e, 0x21, 0x29, 0xc9, 0xa7, 0xa5, 0x7b, 0xd3, 0x18, 0x22,
	0x8b, 0x49, 0xe4, 0x26, 0x91, 0x45, 0xca, 0xe0, 0x54, 0x69, 0xa1, 0xb8, 0x16, 0xc9, 0x20, 0x25,
	0xf9, 0x64, 0x75, 0xce, 0x7c, 0x0a, 0x0b, 0x29, 0xec, 0x56, 0x1e, 0xca, 0x00, 0xd1, 0x4b, 0x18,
	0x6d, 0xb1, 0xef, 0x5b, 0x9b, 0x0c, 0x7f, 0xc1, 0x3f, 0x98, 0xec, 0x06, 0x60, 0xf3, 0xcc, 0x5b,
	0xe9, 0xfb, 0x2c, 0xe1, 0xa4, 0x6b, 0xe5, 0xce, 0x24, 0x24, 0x1d, 0xe4, 0x93, 0xd5, 0x9c, 0x85,
	0xcb, 0xd8, 0xf7, 0xea, 0xa5, 0x07, 0xb3, 0x35, 0xc4, 0x85, 0x0f, 0x2e, 0xc5, 0xcb, 0x5e, 0x18,
	0x4b, 0x97, 0x30, 0x56, 0x1a, 0x15, 0x1a, 0xde, 0xb9, 0xbb, 0x8e, 0x35, 0xf8, 0xa4, 0xb2, 0x07,
	0x38, 0xdb, 0xb8, 0x36, 0x61, 0x85, 0x57, 0x40, 0x7e, 0x52, 0x10, 0xfd, 0x43, 0x41, 0x56, 0xc0,
	0xac, 0xd0, 0xa8, 0x78, 0xc3, 0xad, 0x38, 0xb6, 0xf3, 0x4b, 0x53, 0xf4, 0xb7, 0xa6, 0xf5, 0xf4,
	0x11, 0xd8, 0x75, 0x90, 0x51, 0x8d, 0x1c, 0x73, 0xf5, 0x1e, 0x00, 0x00, 0xff, 0xff, 0x4d, 0xe0,
	0xc9, 0x2e, 0x09, 0x02, 0x00, 0x00,
}
