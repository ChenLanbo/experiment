package raft
import (
	"sync"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"time"
	"log"
	"sync/atomic"
)

type Follower struct {
	nodeMaster *NodeMaster
	expire chan bool
	stopped int32
}

func (follower *Follower) Run() {
	if follower.Stopped() {
		return
	}
	processor := NewFollowerRequestProcessor(follower)
	processor.Start()

	expire := <- follower.expire
	if expire {
		// ignore
	}

	processor.Stop()

	follower.nodeMaster.state = CANDIDATE
	log.Println("Change to candidate")
}

func (follower *Follower) Stop() {
	atomic.StoreInt32(&follower.stopped, 1)
	follower.expire <- false
}

func (follower *Follower) Stopped() bool {
	return atomic.LoadInt32(&follower.stopped) == 1
}

///////////////////////////////////////////////////////////////
// FollowerRequestProcessor
///////////////////////////////////////////////////////////////

type FollowerRequestProcessor struct {
	follower *Follower
	stopped int32
	mu sync.Mutex
}

func (processor *FollowerRequestProcessor) Start() {
	go func () {
		processor.mu.Lock()
		defer processor.mu.Unlock()

		for !processor.Stopped() {
			processor.ProcessOnce()
		}
	} ()
}

func (processor *FollowerRequestProcessor) ProcessOnce() {
	queue := processor.follower.nodeMaster.OpsQueue
	store := processor.follower.nodeMaster.store
	success := false
	op := queue.Pull(time.Millisecond * time.Duration(*queuePullTimeout))
	if op == nil {
		// Notify follower
		log.Println(
			processor.follower.nodeMaster.MyEndpoint(),
			"not hear from leader",
			processor.follower.nodeMaster.votedLeader)
		processor.follower.expire <- true
		processor.Stop()
		return
	}

	if op.Request.AppendRequest != nil {
		reply := &pb.AppendReply{}

		if store.CurrentTerm() <= op.Request.AppendRequest.GetTerm() {
			store.SetCurrentTerm(op.Request.AppendRequest.GetTerm())
			processor.follower.nodeMaster.votedLeader = op.Request.AppendRequest.GetLeaderId()

			if store.Match(op.Request.AppendRequest.GetPrevLogIndex(), op.Request.AppendRequest.GetPrevLogTerm()) {
				tempLogs := make([]pb.Log, len(op.Request.AppendRequest.GetLogs()))
				for i := 0; i < len(tempLogs); i++ {
					tempLogs[i] = *op.Request.AppendRequest.GetLogs()[i]
				}

				err := store.Append(op.Request.AppendRequest.GetCommitIndex(), tempLogs)
				if err == nil {
					success = true
				}
			} else {
				log.Println("Previous log index and term from leader doesn't match the follower's log")
			}
		}

		reply.Success = proto.Bool(success)
		reply.Term = proto.Uint64(store.CurrentTerm())
		op.Callback <- *NewRaftReply(nil, reply, nil)

	} else if op.Request.VoteRequest != nil {
		reply := &pb.VoteReply{}

		if store.CurrentTerm() == op.Request.VoteRequest.GetTerm() {
			if processor.follower.nodeMaster.votedLeader != "" &&
			   processor.follower.nodeMaster.votedLeader == op.Request.VoteRequest.GetCandidateId() &&
			   store.Match(op.Request.VoteRequest.GetLastLogIndex(), op.Request.VoteRequest.GetLastLogTerm()) {
				success = true
			}
		} else if store.CurrentTerm() < op.Request.VoteRequest.GetTerm() {
			if store.Match(op.Request.VoteRequest.GetLastLogIndex(), op.Request.VoteRequest.GetLastLogTerm()) {
				success = true
			}
		}

		if success {
			store.SetCurrentTerm(op.Request.VoteRequest.GetTerm())
		}

		reply.Granted = proto.Bool(success)
		reply.Term = proto.Uint64(store.CurrentTerm())
		op.Callback <- *NewRaftReply(reply, nil, nil)

	} else if op.Request.PutRequest != nil {
		reply := &pb.PutReply{
			Success:proto.Bool(success),
			LeaderId:proto.String(processor.follower.nodeMaster.votedLeader)}
		op.Callback <- *NewRaftReply(nil, nil, reply)
	}
}

func (processor *FollowerRequestProcessor) Stop() {
	atomic.StoreInt32(&processor.stopped, 1)
}

func (processor *FollowerRequestProcessor) Stopped() bool {
	return atomic.LoadInt32(&processor.stopped) == 1
}

///////////////////////////////////////////////////////////////
// Constructors
///////////////////////////////////////////////////////////////

func NewFollowerRequestProcessor(follower *Follower) *FollowerRequestProcessor {
	processor := &FollowerRequestProcessor{}
	processor.follower = follower
	return processor
}

func NewFollower(nodeMaster *NodeMaster) (*Follower) {
	if nodeMaster == nil {
		return nil
	}

	follower := &Follower{}
	follower.nodeMaster = nodeMaster
	follower.expire = make(chan bool, 2)
	follower.stopped = 0

	return follower
}