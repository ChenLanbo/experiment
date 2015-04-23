// Code generated by protoc-gen-go.
// source: internal.proto
// DO NOT EDIT!

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	internal.proto

It has these top-level messages:
	VoteRequest
	VoteReply
	AppendRequest
	AppendReply
	Log
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = math.Inf

// next tag: 5
type VoteRequest struct {
	Term             *uint64 `protobuf:"varint,1,req,name=term" json:"term,omitempty"`
	CandidateId      *string `protobuf:"bytes,2,req,name=candidateId" json:"candidateId,omitempty"`
	LastLogIndex     *uint64 `protobuf:"varint,3,req,name=lastLogIndex" json:"lastLogIndex,omitempty"`
	LastLogTerm      *uint64 `protobuf:"varint,4,req,name=lastLogTerm" json:"lastLogTerm,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *VoteRequest) Reset()         { *m = VoteRequest{} }
func (m *VoteRequest) String() string { return proto1.CompactTextString(m) }
func (*VoteRequest) ProtoMessage()    {}

func (m *VoteRequest) GetTerm() uint64 {
	if m != nil && m.Term != nil {
		return *m.Term
	}
	return 0
}

func (m *VoteRequest) GetCandidateId() string {
	if m != nil && m.CandidateId != nil {
		return *m.CandidateId
	}
	return ""
}

func (m *VoteRequest) GetLastLogIndex() uint64 {
	if m != nil && m.LastLogIndex != nil {
		return *m.LastLogIndex
	}
	return 0
}

func (m *VoteRequest) GetLastLogTerm() uint64 {
	if m != nil && m.LastLogTerm != nil {
		return *m.LastLogTerm
	}
	return 0
}

// next tag: 3
type VoteReply struct {
	CurrentTerm      *uint64 `protobuf:"varint,1,req,name=currentTerm" json:"currentTerm,omitempty"`
	Granted          *bool   `protobuf:"varint,2,req,name=granted" json:"granted,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *VoteReply) Reset()         { *m = VoteReply{} }
func (m *VoteReply) String() string { return proto1.CompactTextString(m) }
func (*VoteReply) ProtoMessage()    {}

func (m *VoteReply) GetCurrentTerm() uint64 {
	if m != nil && m.CurrentTerm != nil {
		return *m.CurrentTerm
	}
	return 0
}

func (m *VoteReply) GetGranted() bool {
	if m != nil && m.Granted != nil {
		return *m.Granted
	}
	return false
}

type AppendRequest struct {
	Term             *uint64 `protobuf:"varint,1,req,name=term" json:"term,omitempty"`
	LeaderId         *string `protobuf:"bytes,2,req,name=leaderId" json:"leaderId,omitempty"`
	PrevLogIndex     *uint64 `protobuf:"varint,3,req,name=prevLogIndex" json:"prevLogIndex,omitempty"`
	PrevLogTerm      *uint64 `protobuf:"varint,4,req,name=prevLogTerm" json:"prevLogTerm,omitempty"`
	CommitIndex      *uint64 `protobuf:"varint,5,req,name=commitIndex" json:"commitIndex,omitempty"`
	Logs             []*Log  `protobuf:"bytes,6,rep,name=logs" json:"logs,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *AppendRequest) Reset()         { *m = AppendRequest{} }
func (m *AppendRequest) String() string { return proto1.CompactTextString(m) }
func (*AppendRequest) ProtoMessage()    {}

func (m *AppendRequest) GetTerm() uint64 {
	if m != nil && m.Term != nil {
		return *m.Term
	}
	return 0
}

func (m *AppendRequest) GetLeaderId() string {
	if m != nil && m.LeaderId != nil {
		return *m.LeaderId
	}
	return ""
}

func (m *AppendRequest) GetPrevLogIndex() uint64 {
	if m != nil && m.PrevLogIndex != nil {
		return *m.PrevLogIndex
	}
	return 0
}

func (m *AppendRequest) GetPrevLogTerm() uint64 {
	if m != nil && m.PrevLogTerm != nil {
		return *m.PrevLogTerm
	}
	return 0
}

func (m *AppendRequest) GetCommitIndex() uint64 {
	if m != nil && m.CommitIndex != nil {
		return *m.CommitIndex
	}
	return 0
}

func (m *AppendRequest) GetLogs() []*Log {
	if m != nil {
		return m.Logs
	}
	return nil
}

type AppendReply struct {
	Term             *uint64 `protobuf:"varint,1,req,name=term" json:"term,omitempty"`
	Success          *bool   `protobuf:"varint,2,req,name=success" json:"success,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *AppendReply) Reset()         { *m = AppendReply{} }
func (m *AppendReply) String() string { return proto1.CompactTextString(m) }
func (*AppendReply) ProtoMessage()    {}

func (m *AppendReply) GetTerm() uint64 {
	if m != nil && m.Term != nil {
		return *m.Term
	}
	return 0
}

func (m *AppendReply) GetSuccess() bool {
	if m != nil && m.Success != nil {
		return *m.Success
	}
	return false
}

// next tag: 6
type Log struct {
	LogId            *uint64  `protobuf:"varint,1,req,name=logId" json:"logId,omitempty"`
	Records          []string `protobuf:"bytes,2,rep,name=records" json:"records,omitempty"`
	ReadSet          []int32  `protobuf:"varint,3,rep,name=readSet" json:"readSet,omitempty"`
	WriteSet         []int32  `protobuf:"varint,4,rep,name=writeSet" json:"writeSet,omitempty"`
	Checksum         *uint32  `protobuf:"varint,5,req,name=checksum" json:"checksum,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *Log) Reset()         { *m = Log{} }
func (m *Log) String() string { return proto1.CompactTextString(m) }
func (*Log) ProtoMessage()    {}

func (m *Log) GetLogId() uint64 {
	if m != nil && m.LogId != nil {
		return *m.LogId
	}
	return 0
}

func (m *Log) GetRecords() []string {
	if m != nil {
		return m.Records
	}
	return nil
}

func (m *Log) GetReadSet() []int32 {
	if m != nil {
		return m.ReadSet
	}
	return nil
}

func (m *Log) GetWriteSet() []int32 {
	if m != nil {
		return m.WriteSet
	}
	return nil
}

func (m *Log) GetChecksum() uint32 {
	if m != nil && m.Checksum != nil {
		return *m.Checksum
	}
	return 0
}

func init() {
}