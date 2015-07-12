package raft

import (
	"testing"
	"github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"time"
)

func TestOpsQueue(t *testing.T) {
    queue := NewOperationQueue()

	voteRequest := &protos.VoteRequest{Term:proto.Uint64(2),
									   CandidateId:proto.String("localhost:12345"),
									   LastLogTerm:proto.Uint64(1),
									   LastLogIndex:proto.Uint64(1)}

    voteReply := &protos.VoteReply{Granted:proto.Bool(true),
                                   Term:proto.Uint64(1)}

	op := NewRaftOperation(NewRaftRequest(voteRequest, nil))

	queue.Push(op)

	pulledOp := queue.Pull(time.Millisecond * 1000)

    if pulledOp == nil {
        t.Fatal("Pulled nil operation")
        t.Fail()
    }
	if pulledOp != op {
        t.Fatal("Invalid operation")
        t.Fail()
	}


    t.Log("Send reply")
    pulledOp.Callback <- *NewRaftReply(voteReply, nil)

    reply := <-pulledOp.Callback
    if reply.VoteReply == nil {
        t.Fatal("Nil reply")
        t.Fail()
    }
    if !*reply.VoteReply.Granted {
        t.Fatal("Unexpected vote reply")
        t.Fail()
    }
}
