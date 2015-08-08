package raft

import (
	pb "github.com/chenlanbo/experiment/paxos/protos"
)

type RaftRequest struct {
	VoteRequest *pb.VoteRequest
	AppendRequest *pb.AppendRequest
	PutRequest *pb.PutRequest
}

type RaftReply struct {
	VoteReply *pb.VoteReply
	AppendReply *pb.AppendReply
	PutReply *pb.PutReply
}

type RaftOperation struct {
	Request *RaftRequest
	Callback chan RaftReply
}

func NewRaftRequest(voteRequest *pb.VoteRequest, appendRequest *pb.AppendRequest, putRequest *pb.PutRequest) (*RaftRequest) {
	return &RaftRequest{
		VoteRequest:voteRequest,
		AppendRequest:appendRequest,
		PutRequest:putRequest}
}

func NewRaftReply(voteReply *pb.VoteReply, appendReply *pb.AppendReply, putReply *pb.PutReply) (*RaftReply) {
	return &RaftReply{
		VoteReply:voteReply,
		AppendReply:appendReply,
		PutReply:putReply}
}

func NewRaftOperation(raftRequest *RaftRequest) (*RaftOperation) {
	if raftRequest == nil {
		return nil
	}

	return &RaftOperation{Request:raftRequest,
					      Callback: make(chan RaftReply, 1)}
}