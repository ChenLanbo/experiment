package raft

import (
	pb "github.com/chenlanbo/experiment/paxos/protos"
)

type RaftRequest struct {
	VoteRequest *pb.VoteRequest
	AppendRequest *pb.AppendRequest
}

type RaftReply struct {
	VoteReply *pb.VoteReply
	AppendReply *pb.AppendReply
}

type RaftOperation struct {
	Request *RaftRequest
	Callback chan RaftReply
}

func NewRaftRequest(voteRequest *pb.VoteRequest, appendRequest *pb.AppendRequest) (*RaftRequest) {
	if voteRequest == nil && appendRequest == nil {
		return nil
	}
	if voteRequest != nil && appendRequest != nil {
		return nil
	}

	return &RaftRequest{VoteRequest:voteRequest,
						AppendRequest:appendRequest}
}

func NewRaftReply(voteReply *pb.VoteReply, appendReply *pb.AppendReply) (*RaftReply) {
	if voteReply == nil && appendReply == nil {
		return nil
	}
	if voteReply != nil && appendReply != nil {
		return nil
	}

	return &RaftReply{VoteReply:voteReply,
					  AppendReply:appendReply}
}

func NewRaftOperation(raftRequest *RaftRequest) (*RaftOperation) {
	if raftRequest == nil {
		return nil
	}

	return &RaftOperation{Request:raftRequest,
					      Callback: make(chan RaftReply, 1)}
}