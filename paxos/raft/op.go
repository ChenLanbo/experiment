package raft

import (
	pb "github.com/chenlanbo/experiment/paxos/protos"
)

type RaftRequest struct {
	VoteRequest *pb.VoteRequest
	AppendRequest *pb.AppendRequest
	PutRequest *pb.PutRequest
	GetRequest *pb.GetRequest
}

type RaftReply struct {
	VoteReply *pb.VoteReply
	AppendReply *pb.AppendReply
	PutReply *pb.PutReply
	GetReply *pb.GetReply
}

type RaftOperation struct {
	Request *RaftRequest
	Callback chan RaftReply
}

func EmptyRaftRequest() *RaftRequest {
	return &RaftRequest{}
}

func WithVoteRequest(raftRequest *RaftRequest, voteRequest *pb.VoteRequest) *RaftRequest {
	raftRequest.VoteRequest = voteRequest
	return raftRequest
}

func WithAppendRequest(raftRequest *RaftRequest, appendRequest *pb.AppendRequest) *RaftRequest {
	raftRequest.AppendRequest = appendRequest
	return raftRequest
}

func WithPutRequest(raftRequest *RaftRequest, putRequest *pb.PutRequest) *RaftRequest {
	raftRequest.PutRequest = putRequest
	return raftRequest
}

func WithGetRequest(raftRequest *RaftRequest, getRequest *pb.GetRequest) *RaftRequest {
	raftRequest.GetRequest = getRequest
	return raftRequest
}

func EmptyRaftReply() *RaftReply {
	return &RaftReply{}
}

func WithVoteReply(raftReply *RaftReply, voteReply *pb.VoteReply) *RaftReply {
	raftReply.VoteReply = voteReply
	return raftReply
}

func WithAppendReply(raftReply *RaftReply, appendReply *pb.AppendReply) *RaftReply {
	raftReply.AppendReply = appendReply
	return raftReply
}

func WithPutReply(raftReply *RaftReply, putReply *pb.PutReply) *RaftReply {
	raftReply.PutReply = putReply
	return raftReply
}

func WithGetReply(raftReply *RaftReply, getReply *pb.GetReply) *RaftReply {
	raftReply.GetReply = getReply
	return raftReply
}

func NewRaftOperation(raftRequest *RaftRequest) (*RaftOperation) {
	if raftRequest == nil {
		return nil
	}

	return &RaftOperation{Request:raftRequest,
					      Callback: make(chan RaftReply, 1)}
}