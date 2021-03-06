package raft
import (
	"sync"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"time"
	"log"
	"sync/atomic"
	"golang.org/x/net/context"
)

type Follower struct {
	nodeMaster *NodeMaster

	expireCtx  context.Context
	expireCancel context.CancelFunc

	stopped    int32
	stopCtx    context.Context
	stopCancel context.CancelFunc
}

func (follower *Follower) Run() {
	if follower.Stopped() {
		return
	}
	processor := NewFollowerRequestProcessor(follower)
	processor.Start()


	select {
	case <-follower.stopCtx.Done():
		log.Println("Stop processor")
		processor.Stop()
	case <-follower.expireCtx.Done():
		processor.Stop()
		follower.nodeMaster.state = CANDIDATE
		log.Println("Change to candidate")
	}
}

func (follower *Follower) Stop() {
	atomic.StoreInt32(&follower.stopped, 1)
	follower.stopCancel()
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
	lastSawAppend int64
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

	if op != nil && op.Request.AppendRequest != nil {

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


		reply := &pb.AppendReply{
			Success:proto.Bool(success),
			Term:proto.Uint64(store.CurrentTerm())}
		op.Callback <- *WithAppendReply(EmptyRaftReply(), reply)

	} else if op != nil && op.Request.VoteRequest != nil {

		// If votedLeader is null or candidateId, and candidate's log is at least as up-to-date
		// as receiver’s log, grant vote.
		if store.CurrentTerm() == op.Request.VoteRequest.GetTerm() {
			if processor.follower.nodeMaster.votedLeader == "" &&
			   store.OtherLogUpToDate(op.Request.VoteRequest.GetLastLogIndex(), op.Request.VoteRequest.GetLastLogTerm()) {
				success = true
				processor.follower.nodeMaster.votedLeader = op.Request.VoteRequest.GetCandidateId()
			}

			if processor.follower.nodeMaster.votedLeader != "" &&
			   processor.follower.nodeMaster.votedLeader == op.Request.VoteRequest.GetCandidateId() &&
			   store.OtherLogUpToDate(op.Request.VoteRequest.GetLastLogIndex(), op.Request.VoteRequest.GetLastLogTerm()) {
				success = true
				processor.follower.nodeMaster.votedLeader = op.Request.VoteRequest.GetCandidateId()
			}

		} else if store.CurrentTerm() < op.Request.VoteRequest.GetTerm() {
			store.SetCurrentTerm(op.Request.VoteRequest.GetTerm())
			if store.OtherLogUpToDate(op.Request.VoteRequest.GetLastLogIndex(), op.Request.VoteRequest.GetLastLogTerm()) {
				success = true
				processor.follower.nodeMaster.votedLeader = op.Request.VoteRequest.GetCandidateId()
			} else {
				processor.follower.nodeMaster.votedLeader = ""
			}
		}

		reply := &pb.VoteReply{
			Granted:proto.Bool(success),
			Term:proto.Uint64(store.CurrentTerm())}
		op.Callback <- *WithVoteReply(EmptyRaftReply(), reply)

	} else if op != nil && op.Request.PutRequest != nil {

		reply := &pb.PutReply{
			Success:proto.Bool(success),
			LeaderId:proto.String(processor.follower.nodeMaster.votedLeader)}
		op.Callback <- *WithPutReply(EmptyRaftReply(), reply)

	} else if op != nil && op.Request.GetRequest != nil {

		reply := &pb.GetReply{
			Success:proto.Bool(success),
			LeaderId:proto.String(processor.follower.nodeMaster.votedLeader),
			Key:proto.String(op.Request.GetRequest.GetKey())}
		op.Callback <- *WithGetReply(EmptyRaftReply(), reply)
	}

	// Check election timeout of leader
	if op != nil && op.Request.AppendRequest != nil {
		processor.lastSawAppend = time.Now().UnixNano()
	} else if time.Now().UnixNano() - processor.lastSawAppend >= int64(time.Second) {
		// Election timeout (not hear append from leader), notify follower to become candidate
		log.Println(
			processor.follower.nodeMaster.MyEndpoint(),
			"not hear from leader",
			processor.follower.nodeMaster.votedLeader)
		processor.follower.expireCancel()
		processor.Stop()
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
	processor.lastSawAppend = time.Now().UnixNano()
	return processor
}

func NewFollower(nodeMaster *NodeMaster) (*Follower) {
	if nodeMaster == nil {
		return nil
	}

	follower := &Follower{}
	follower.nodeMaster = nodeMaster

	follower.stopped = 0
	follower.stopCtx, follower.stopCancel = context.WithCancel(context.Background())

	follower.expireCtx, follower.expireCancel = context.WithCancel(context.Background())

	return follower
}