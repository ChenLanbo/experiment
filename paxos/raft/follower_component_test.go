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

func TestFollowerTimeoutWithNoRequests(t *testing.T) {
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

func TestFollowerHandleVote(t *testing.T) {
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

	// Another candidate send vote in the same term
	voteRequest2 := &pb.VoteRequest{
		Term:proto.Uint64(1),
		CandidateId:proto.String(tt.peers[2]),
		LastLogTerm:proto.Uint64(0),
		LastLogIndex:proto.Uint64(0)}
	op2 := NewRaftOperation(NewRaftRequest(voteRequest2, nil, nil))
	defer close(op2.Callback)
	tt.nodeMaster.OpsQueue.Push(op2)

	tt.processor.ProcessOnce()
	reply = <- op2.Callback
	if reply.VoteReply == nil {
		t.Fatal("Empty vote reply")
		t.Fail()
	}
	if reply.VoteReply.GetGranted() {
		t.Fatal("Vote should not be granted")
		t.Fail()
	}
}

func TestFollowerHandleVoteLogNotUpToDate(t *testing.T) {
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
		t.Fail()
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
		t.Fail()
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
		t.Fail()
	}
}

func TestFollowerHandleAppend(t *testing.T) {
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
		t.Fatal("Follower's term not advanced")
		t.Fail()
	}
	if tt.nodeMaster.store.LatestIndex() != 2 {
		t.Fatal("Follower's store rejected new logs")
		t.Fail()
	}

	reply := <- op.Callback
	if reply.AppendReply == nil {
		t.Fatal("Empty append reply")
		t.Fail()
	}
	if !reply.AppendReply.GetSuccess() {
		t.Fatal("Append failed")
		t.Fail()
	}
}

func TestReplicateFromPreviousLog(t *testing.T) {
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
	if *reply.AppendReply.Success {
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

func TestVoteAndAppend(t *testing.T) {
	// TODO(lanbochen): fill this test
}