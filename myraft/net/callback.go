package net

import myproto "github.com/chenlanbo/experiment/myraft/proto"

type VoteCallback struct {
    Req *myproto.VoteRequest
    Rep chan myproto.VoteReply
}

type AppendCallback struct {
    Req *myproto.AppendRequest
    Rep chan myproto.AppendReply
}

