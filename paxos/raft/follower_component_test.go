package raft
import (
	"testing"

	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"github.com/golang/mock/gomock"
	"github.com/chenlanbo/experiment/paxos/raft/test"
	"time"
)

type FollowerComponentTest struct {
	mockCtrl *gomock.Controller
	mockExchange *test.MockMessageExchange

	peers []string
	me int

	nodeMaster *NodeMaster
	follower   *Follower
	processor  *FollowerRequestProcessor
}

func (tt *FollowerComponentTest) setUp(t *testing.T) {
	tt.mockCtrl = gomock.NewController(t)
	tt.mockExchange = test.NewMockMessageExchange(tt.mockCtrl)

	tt.peers = []string{"localhost:10001", "localhost:10002", "localhost:10003"}
	tt.me = 0

	tt.nodeMaster = NewNodeMaster(tt.mockExchange, tt.peers, tt.me)
	tt.follower = NewFollower(tt.nodeMaster)
	tt.processor = NewFollowerRequestProcessor(tt.follower)
}

func (tt *FollowerComponentTest) tearDown(t *testing.T) {
	tt.processor.Stop()
	tt.nodeMaster.Stop()
	tt.follower = nil
	tt.nodeMaster = nil

	tt.mockCtrl.Finish()
	tt.mockExchange = nil
	tt.mockCtrl = nil
}

func TestFollowerProcessorLeaderElectionTimeout(t *testing.T) {
	tt := &FollowerComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	tt.processor.ProcessOnce()

	select {
	case <- tt.follower.expireCtx.Done():
		// Success
		t.Log("Timeout expected\n")
	case <- time.After(time.Second * 2):
		t.Fail()
	}
}

func TestFollowerProcessorHandleVote(t *testing.T) {
	tt := &FollowerComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	voteRequest1 := &pb.VoteRequest{
		Term:proto.Uint64(1),
		CandidateId:proto.String(tt.peers[1]),
		LastLogTerm:proto.Uint64(0),
		LastLogIndex:proto.Uint64(0)}
	op1 := NewRaftOperation(NewRaftRequest(voteRequest1, nil, nil))
	defer close(op1.Callback)
	tt.nodeMaster.OpsQueue.Push(op1)

	tt.processor.ProcessOnce()
	if tt.nodeMaster.store.CurrentTerm() != 1 {
		t.Fail()
	}

	reply := <- op1.Callback
	if reply.VoteReply == nil {
		t.Fatal("Empty vote reply")
		t.Fail()
	}
	if !reply.VoteReply.GetGranted() {
		t.Fatal("Vote not granted")
		t.Fail()
	}

	// The same vote request should also succeed
	op2 := NewRaftOperation(NewRaftRequest(voteRequest1, nil, nil))
	defer close(op2.Callback)
	tt.nodeMaster.OpsQueue.Push(op2)

	tt.processor.ProcessOnce()

	reply = <- op2.Callback
	if reply.VoteReply == nil {
		t.Fatal("Empty vote reply")
		t.Fail()
	}
	if !reply.VoteReply.GetGranted() {
		t.Fatal("Vote not granted")
		t.Fail()
	}

	// Another candidate send vote in the same term, already elected a candidate
	voteRequest2 := &pb.VoteRequest{
		Term:proto.Uint64(1),
		CandidateId:proto.String(tt.peers[2]),
		LastLogTerm:proto.Uint64(0),
		LastLogIndex:proto.Uint64(0)}
	op3 := NewRaftOperation(NewRaftRequest(voteRequest2, nil, nil))
	defer close(op3.Callback)
	tt.nodeMaster.OpsQueue.Push(op3)

	tt.processor.ProcessOnce()
	reply = <- op3.Callback
	if reply.VoteReply == nil {
		t.Error("Empty vote reply")
	}
	if reply.VoteReply.GetGranted() {
		t.Error("Vote should not be granted")
	}
}

func TestFollowerProcessorHandleVoteLogNotUpToDate(t *testing.T) {
	tt := &FollowerComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	tt.nodeMaster.store.IncrementCurrentTerm()
	tt.nodeMaster.store.Write([]byte("abc"))
	tt.nodeMaster.store.Write([]byte("abc"))

	// Voter's log not up to date (index behind)
	voteRequest1 := &pb.VoteRequest{
		Term:proto.Uint64(2),
		CandidateId:proto.String(tt.peers[2]),
		LastLogTerm:proto.Uint64(1),
		LastLogIndex:proto.Uint64(1)}
	op1 := NewRaftOperation(NewRaftRequest(voteRequest1, nil, nil))
	defer close(op1.Callback)
	tt.nodeMaster.OpsQueue.Push(op1)

	tt.processor.ProcessOnce()
	reply := <- op1.Callback
	if reply.VoteReply == nil || reply.VoteReply.GetGranted() {
		t.Error("Voter's log not up to date, should not grant")
	}

	// Voter's log up to date
	voteRequest2 := &pb.VoteRequest{
		Term:proto.Uint64(2),
		CandidateId:proto.String(tt.peers[2]),
		LastLogTerm:proto.Uint64(1),
		LastLogIndex:proto.Uint64(2)}
	op2 := NewRaftOperation(NewRaftRequest(voteRequest2, nil, nil))
	defer close(op2.Callback)
	tt.nodeMaster.OpsQueue.Push(op2)

	tt.processor.ProcessOnce()
	reply = <- op2.Callback
	if reply.VoteReply == nil || !reply.VoteReply.GetGranted() {
		t.Error("Voter's log up to date, should grant")
	}

	tt.nodeMaster.store.IncrementCurrentTerm()
	tt.nodeMaster.store.Write([]byte("abc"))

	// Voter's log not up to date (term behind)
	voteRequest3 := &pb.VoteRequest{
		Term:proto.Uint64(3),
		CandidateId:proto.String(tt.peers[2]),
		LastLogTerm:proto.Uint64(1),
		LastLogIndex:proto.Uint64(3)}
	op3 := NewRaftOperation(NewRaftRequest(voteRequest3, nil, nil))
	defer close(op3.Callback)
	tt.nodeMaster.OpsQueue.Push(op3)

	tt.processor.ProcessOnce()
	reply = <- op3.Callback
	if reply.VoteReply == nil || reply.VoteReply.GetGranted() {
		t.Error("Voter's log not up to date, should not grant")
	}
}

func TestFollowerProcessorHandleAppend(t *testing.T) {
	tt := &FollowerComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	appendRequest := &pb.AppendRequest{
		Term:proto.Uint64(1),
		LeaderId:proto.String(tt.peers[1]),
		PrevLogIndex:proto.Uint64(0),
		PrevLogTerm:proto.Uint64(0),
		CommitIndex:proto.Uint64(0),
		Logs:[]*pb.Log{
			&pb.Log{Term:proto.Uint64(1), LogId:proto.Uint64(1)},
			&pb.Log{Term:proto.Uint64(1), LogId:proto.Uint64(2)}}}

	op := NewRaftOperation(NewRaftRequest(nil, appendRequest, nil))
	tt.nodeMaster.OpsQueue.Push(op)

	tt.processor.ProcessOnce()
	if tt.nodeMaster.store.CurrentTerm() != 1 {
		t.Error("Follower's term not advanced")
	}
	if tt.nodeMaster.store.LatestIndex() != 2 {
		t.Error("Follower's store rejected new logs")
	}

	reply := <- op.Callback
	if reply.AppendReply == nil {
		t.Error("Empty append reply")
	}
	if !reply.AppendReply.GetSuccess() {
		t.Error("Append failed")
	}
}

func TestFollowerProcessorReplicateFromPreviousLog(t *testing.T) {
	tt := FollowerComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	fakeData := make([]byte, 0)
	n := 4
	for i := 0; i < n; i++ {
		tt.nodeMaster.store.IncrementCurrentTerm()
		tt.nodeMaster.store.Write(fakeData)
	}

	// PrevLogIndex and PrevLogTerm do not match
	newTerm := tt.nodeMaster.store.CurrentTerm() + 1
	appendRequest1 := &pb.AppendRequest{
		Term:proto.Uint64(newTerm),
		LeaderId:proto.String(tt.peers[1]),
		PrevLogIndex:proto.Uint64(4),
		PrevLogTerm:proto.Uint64(newTerm - 2),
		CommitIndex:proto.Uint64(0),
		Logs:[]*pb.Log{
			&pb.Log{Term:proto.Uint64(newTerm), LogId:proto.Uint64(uint64(n + 1))},
			&pb.Log{Term:proto.Uint64(newTerm), LogId:proto.Uint64(uint64(n + 2))}}}

	op1 := NewRaftOperation(NewRaftRequest(nil, appendRequest1, nil))
	tt.nodeMaster.OpsQueue.Push(op1)

	tt.processor.ProcessOnce()
	reply := <- op1.Callback
	if reply.AppendReply.GetSuccess() {
		t.Fail()
	}

	// Use earlier PrevLogIndex and PrevLogTerm
	appendRequest2 := &pb.AppendRequest{
		Term:proto.Uint64(newTerm),
		LeaderId:proto.String(tt.peers[1]),
		PrevLogIndex:proto.Uint64(3),
		PrevLogTerm:proto.Uint64(newTerm - 2),
		CommitIndex:proto.Uint64(0),
		Logs:[]*pb.Log{
			&pb.Log{Term:proto.Uint64(newTerm), LogId:proto.Uint64(uint64(n))},
			&pb.Log{Term:proto.Uint64(newTerm), LogId:proto.Uint64(uint64(n + 1))}}}

	op2 := NewRaftOperation(NewRaftRequest(nil, appendRequest2, nil))
	tt.nodeMaster.OpsQueue.Push(op2)

	tt.processor.ProcessOnce()
	reply = <- op2.Callback
	if !reply.AppendReply.GetSuccess() {
		t.Fail()
	}
	if tt.nodeMaster.store.LatestIndex() != uint64(n + 1) {
		t.Fail()
	}
	l := tt.nodeMaster.store.Read(uint64(n))
	if l.GetTerm() != newTerm || l.GetLogId() != uint64(n) {
		t.Fail()
	}
}

func TestFollowerProcessorHandlePut(t *testing.T) {
	tt := FollowerComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	putRequest := &pb.PutRequest{
		Key:proto.String("abc"), Value:[]byte("abc")}
	op := NewRaftOperation(NewRaftRequest(nil, nil, putRequest))
	defer close(op.Callback)
	tt.nodeMaster.OpsQueue.Push(op)

	tt.processor.ProcessOnce()
	reply := <- op.Callback
	if reply.PutReply == nil || reply.PutReply.GetSuccess() {
		t.Error("Put should not succeed")
	}
}

func TestFollowerChangeToCandidate(t *testing.T) {
	tt := FollowerComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	tt.follower.Run()

	if tt.nodeMaster.state != CANDIDATE {
		t.Error("Should have become a candidate")
	}
}

func TestFollowerStop(t *testing.T) {
	tt := FollowerComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	go func() {
		tt.follower.Run()
	} ()

	appendRequest := &pb.AppendRequest{
		Term:proto.Uint64(1),
		LeaderId:proto.String(tt.peers[1]),
		PrevLogIndex:proto.Uint64(0),
		PrevLogTerm:proto.Uint64(0),
		CommitIndex:proto.Uint64(0)}
	for i := 0; i < 10; i++ {
		op := NewRaftOperation(NewRaftRequest(nil, appendRequest, nil))
		defer close(op.Callback)
		tt.nodeMaster.OpsQueue.Push(op)
		reply := <- op.Callback
		if reply.AppendReply == nil {
			t.Fail()
		}
	}

	tt.follower.Stop()
	time.Sleep(time.Second)
	if !tt.processor.follower.Stopped() {
		t.Error("Follower processor should have stopped")
	}
}