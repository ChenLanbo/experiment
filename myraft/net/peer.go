package net

import "github.com/chenlanbo/experiment/myraft/proto"

type PeerClient interface {
    VoteRequest(req *proto.VoteRequest) (*proto.VoteReply, error)
    AppendRequest(req *proto.AppendRequest) (*proto.AppendReply, error)
}

type PeerServer interface {
    Start()
    Shutdown()
    VoteChan() chan *VoteCallback
    AppendChan() chan *AppendCallback
}
