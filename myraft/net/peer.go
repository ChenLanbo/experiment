package net

import "github.com/chenlanbo/experiment/myraft/proto"

type Peer interface {
    SendVoteRequest(req *proto.VoteRequest) (*proto.VoteReply, error)
}
