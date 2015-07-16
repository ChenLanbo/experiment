package raft
import (
	"sync"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"time"
	"log"
)

type Follower struct {
	mu sync.Mutex
	nodeMaster *NodeMaster
}

func (follower *Follower) Run() {
	follower.ProcessRequest()
}

func (follower *Follower) ProcessRequest() {
	go func() {
		for follower.isFollower() {
			follower.ProcessOneRequest()
		}
	} ()
}

func (follower *Follower) ProcessOneRequest() {
	follower.mu.Lock()
	defer follower.mu.Unlock()

	op := follower.nodeMaster.OpsQueue.Pull(time.Millisecond * time.Duration(*queuePullTimeout))

	if op == nil {
		// Timeout elapses without receiving AppendEntries RPC from current leader or granting vote to candidate
		follower.nodeMaster.state = CANDIDATE
		return
	}

	if op.Request.AppendRequest != nil {
		reply := &pb.AppendReply{}

		if follower.nodeMaster.store.CurrentTerm() > *op.Request.AppendRequest.Term {
			reply.Success = proto.Bool(false)
		} else {
			follower.nodeMaster.store.SetCurrentTerm(*op.Request.AppendRequest.Term)

			if !follower.nodeMaster.store.Match(*op.Request.AppendRequest.PrevLogIndex,
											    *op.Request.AppendRequest.PrevLogTerm) {
				log.Fatal("Previous log index and term from leader doesn't match the follower's log")
				reply.Success = proto.Bool(false)
			} else {
				tempLogs := make([]pb.Log, len(op.Request.AppendRequest.Logs))
				for i := 0; i < len(tempLogs); i++ {
					tempLogs[i] = *op.Request.AppendRequest.Logs[i]
				}
				err := follower.nodeMaster.store.Append(*op.Request.AppendRequest.CommitIndex,
												 	    tempLogs)
				if err != nil {
					reply.Success = proto.Bool(false)
				} else {
					reply.Success = proto.Bool(true)
				}
			}
		}

		reply.Term = proto.Uint64(follower.nodeMaster.store.CurrentTerm())
		op.Callback <- *NewRaftReply(nil, reply)
	}

	if op.Request.VoteRequest != nil {
		reply := &pb.VoteReply{}

		if follower.nodeMaster.store.CurrentTerm() > *op.Request.VoteRequest.Term {
			reply.Granted = proto.Bool(false)
		} else if follower.nodeMaster.store.CurrentTerm() == *op.Request.VoteRequest.Term {
			if follower.nodeMaster.votedLeader == "" || follower.nodeMaster.votedLeader != *op.Request.VoteRequest.CandidateId {
				// i have voted another guy in this term
				reply.Granted = proto.Bool(false)
			} else {
				// i have voted the same candidate, and check log up-to-date again
				if follower.nodeMaster.store.Match(*op.Request.VoteRequest.LastLogIndex,
											       *op.Request.VoteRequest.LastLogTerm) {
					reply.Granted = proto.Bool(true)
				} else {
					reply.Granted = proto.Bool(false)
				}
			}
		} else {
			if follower.nodeMaster.store.OtherLogUpToDate(*op.Request.VoteRequest.LastLogIndex,
														  *op.Request.VoteRequest.LastLogTerm) {
				// candidate's log is at least as up-to-date as receiver’s log
				follower.nodeMaster.store.SetCurrentTerm(*op.Request.VoteRequest.Term)
				follower.nodeMaster.votedLeader = *op.Request.VoteRequest.CandidateId
				reply.Granted = proto.Bool(true)
			} else {
				reply.Granted = proto.Bool(false)
			}
		}

		reply.Term = proto.Uint64(follower.nodeMaster.store.CurrentTerm())
		op.Callback <- *NewRaftReply(reply, nil)
	}
}

func (follower *Follower) isFollower() bool {
	follower.mu.Lock()
	defer follower.mu.Unlock()

	return follower.nodeMaster.state == FOLLOWER
}

func NewFollower(nodeMaster *NodeMaster) (*Follower) {
	if nodeMaster == nil {
		return nil
	}

	follower := &Follower{}
	follower.nodeMaster = nodeMaster

	return follower
}