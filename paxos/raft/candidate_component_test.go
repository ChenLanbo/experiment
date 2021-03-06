package raft
import (
	"testing"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"github.com/golang/mock/gomock"
	"github.com/chenlanbo/experiment/paxos/raft/test"
	"time"
)


type CandidateComponentTest struct {
	mockCtrl *gomock.Controller
	mockExchange *test.MockMessageExchange

	peers []string
	me int

	nodeMaster *NodeMaster
	candidate *Candidate
}

func (tt *CandidateComponentTest) setUp(t *testing.T, peers []string) {
	tt.mockCtrl = gomock.NewController(t)
	tt.mockExchange = test.NewMockMessageExchange(tt.mockCtrl)

	tt.peers = peers
	tt.me = 0

	tt.nodeMaster = NewNodeMaster(tt.mockExchange, tt.peers, tt.me)
	tt.candidate = NewCandidate(tt.nodeMaster)
	tt.nodeMaster.state = CANDIDATE
}

func (tt *CandidateComponentTest) tearDown(t *testing.T) {
	tt.nodeMaster.Stop()
	tt.candidate = nil
	tt.nodeMaster = nil

	tt.mockCtrl.Finish()
	tt.mockExchange = nil
	tt.mockCtrl = nil
}

func TestCandidateVoteBothPeersGranted(t *testing.T) {
	tt := &CandidateComponentTest{}
	tt.setUp(t, peers1)
	defer tt.tearDown(t)

	reply := &pb.VoteReply{
		Granted:proto.Bool(true),
		Term:proto.Uint64(0)}
	tt.mockExchange.EXPECT().Vote(gomock.Any(), gomock.Any()).Return(reply, nil)

	tt.nodeMaster.store.IncrementCurrentTerm()
	newTerm := tt.nodeMaster.store.CurrentTerm()
	l := tt.nodeMaster.store.Read(tt.nodeMaster.store.LatestIndex())

	voter := NewCandidateVoter(tt.candidate)
	voter.VoteSelfAtTerm(newTerm, l.GetTerm(), l.GetLogId())

	select {
	case result := <- tt.candidate.votedChan:
		if !result {
			t.Fail()
		}
	case <- time.After(time.Second * 10):
		t.Fail()
	}
}

func TestCandidateVoteOnlyOnePeerGranted(t *testing.T) {
	tt := &CandidateComponentTest{}
	tt.setUp(t, peers1)
	defer tt.tearDown(t)

	reply1 := &pb.VoteReply{
		Granted:proto.Bool(false),
		Term:proto.Uint64(0)}
	tt.mockExchange.EXPECT().Vote(gomock.Any(), gomock.Any()).Return(reply1, nil)

	reply2 := &pb.VoteReply{
		Granted:proto.Bool(true),
		Term:proto.Uint64(1)}
	tt.mockExchange.EXPECT().Vote(gomock.Any(), gomock.Any()).Return(reply2, nil)

	tt.nodeMaster.store.IncrementCurrentTerm()
	newTerm := tt.nodeMaster.store.CurrentTerm()
	l := tt.nodeMaster.store.Read(tt.nodeMaster.store.LatestIndex())

	voter := NewCandidateVoter(tt.candidate)
	voter.VoteSelfAtTerm(newTerm, l.GetTerm(), l.GetLogId())

	select {
	case result := <- tt.candidate.votedChan:
		if !result {
			t.Fail()
		}
	case <- time.After(time.Second * 10):
		t.Fail()
	}
}

func TestCandidateVoteNoPeersGranted(t *testing.T) {
	tt := &CandidateComponentTest{}
	tt.setUp(t, peers1)
	defer tt.tearDown(t)

	reply := &pb.VoteReply{
		Granted:proto.Bool(false),
		Term:proto.Uint64(0)}
	tt.mockExchange.EXPECT().Vote(gomock.Any(), gomock.Any()).Return(reply, nil)
	tt.mockExchange.EXPECT().Vote(gomock.Any(), gomock.Any()).Return(reply, nil)

	tt.nodeMaster.store.IncrementCurrentTerm()
	newTerm := tt.nodeMaster.store.CurrentTerm()
	l := tt.nodeMaster.store.Read(tt.nodeMaster.store.LatestIndex())

	voter := NewCandidateVoter(tt.candidate)
	voter.VoteSelfAtTerm(newTerm, l.GetTerm(), l.GetLogId())

	select {
	case result := <- tt.candidate.votedChan:
		if result {
			t.Fail()
		}
	case <- time.After(time.Second * 10):
		t.Fail()
	}
}

func TestCandidateProcessorHandlePut(t *testing.T) {
	tt := &CandidateComponentTest{}
	tt.setUp(t, peers1)
	defer tt.tearDown(t)

	tt.nodeMaster.store.IncrementCurrentTerm()
	newTerm := tt.nodeMaster.store.CurrentTerm()

	processor := NewCandidateRequestProcessor(tt.candidate)
	processor.ProcessRequestsAtTerm(newTerm)

	request := &pb.PutRequest{
		Key:proto.String("abc"),
		Value:[]byte("abc")}
	op := NewRaftOperation(WithPutRequest(EmptyRaftRequest(), request))
	tt.nodeMaster.OpsQueue.Push(op)

	reply := <- op.Callback
	if reply.PutReply == nil || reply.PutReply.GetSuccess() {
		t.Fail()
	}

	processor.Stop()
	if !processor.Stopped() {
		t.Fail()
	}
}

func TestCandidateProcessorHandleVote(t *testing.T) {
	tt := &CandidateComponentTest{}
	tt.setUp(t, peers1)
	defer tt.tearDown(t)

	tt.nodeMaster.store.IncrementCurrentTerm()
	newTerm := tt.nodeMaster.store.CurrentTerm()

	processor := NewCandidateRequestProcessor(tt.candidate)
	processor.ProcessRequestsAtTerm(newTerm)

	request1 := &pb.VoteRequest{
		Term:proto.Uint64(newTerm),
		CandidateId:proto.String(tt.peers[1]),
		LastLogTerm:proto.Uint64(0),
		LastLogIndex:proto.Uint64(0)}
	op1 := NewRaftOperation(WithVoteRequest(EmptyRaftRequest(), request1))
	tt.nodeMaster.OpsQueue.Push(op1)

	reply1 := <- op1.Callback
	if reply1.VoteReply == nil || reply1.VoteReply.GetGranted() {
		t.Fail()
	}

	request2 := &pb.VoteRequest{
		Term:proto.Uint64(newTerm + 1),
		CandidateId:proto.String(tt.peers[1]),
		LastLogTerm:proto.Uint64(0),
		LastLogIndex:proto.Uint64(0)}
	op2 := NewRaftOperation(WithVoteRequest(EmptyRaftRequest(), request2))
	tt.nodeMaster.OpsQueue.Push(op2)

	reply2 := <- op2.Callback
	if reply2.VoteReply == nil || reply2.VoteReply.GetGranted() {
		t.Fail()
	}
	higherTerm := <- tt.candidate.newTermChan
	if higherTerm != newTerm + 1 {
		t.Fail()
	}

	processor.Stop()
	if !processor.Stopped() {
		t.Fail()
	}
}

func TestCandidateProcessorHandleAppend(t *testing.T) {
	tt := &CandidateComponentTest{}
	tt.setUp(t, peers1)
	defer tt.tearDown(t)

	tt.nodeMaster.store.IncrementCurrentTerm()
	newTerm := tt.nodeMaster.store.CurrentTerm()

	processor := NewCandidateRequestProcessor(tt.candidate)
	processor.ProcessRequestsAtTerm(newTerm)

	request := &pb.AppendRequest{
		Term:proto.Uint64(newTerm),
		LeaderId:proto.String(tt.peers[1]),
		PrevLogTerm:proto.Uint64(0),
		PrevLogIndex:proto.Uint64(0),
		Logs:make([]*pb.Log, 0)}
	op := NewRaftOperation(WithAppendRequest(EmptyRaftRequest(), request))
	defer close(op.Callback)
	tt.nodeMaster.OpsQueue.Push(op)

	reply := <- op.Callback
	if reply.AppendReply == nil || reply.AppendReply.GetSuccess() {
		t.Fail()
	}
	higherTerm := <- tt.candidate.newTermChan
	if higherTerm != newTerm {
		t.Fail()
	}

	processor.Stop()
	if !processor.Stopped() {
		t.Fail()
	}
}

func TestCandidateProcessorHandleGet(t *testing.T) {
	tt := &CandidateComponentTest{}
	tt.setUp(t, peers1)
	defer tt.tearDown(t)

	tt.nodeMaster.store.IncrementCurrentTerm()
	newTerm := tt.nodeMaster.store.CurrentTerm()

	processor := NewCandidateRequestProcessor(tt.candidate)
	processor.ProcessRequestsAtTerm(newTerm)

	request := &pb.GetRequest{
		Key:proto.String("abc")}
	op := NewRaftOperation(WithGetRequest(EmptyRaftRequest(), request))
	defer close(op.Callback)
	tt.nodeMaster.OpsQueue.Push(op)

	reply := <- op.Callback
	if reply.GetReply == nil || reply.GetReply.GetSuccess() {
		t.Fail()
	}

	processor.Stop()
}

func TestCandidateRun(t *testing.T) {
	tt := &CandidateComponentTest{}
	tt.setUp(t, peers1)
	defer tt.tearDown(t)

	reply := &pb.VoteReply{
		Granted:proto.Bool(true),
		Term:proto.Uint64(0)}
	tt.mockExchange.EXPECT().Vote(gomock.Any(), gomock.Any()).Return(reply, nil).AnyTimes()

	tt.candidate.Run()

	if tt.nodeMaster.state != LEADER {
		t.Fail()
	}
}

func TestCandidateRun2(t *testing.T) {
	tt := &CandidateComponentTest{}
	tt.setUp(t, peers1)
	defer tt.tearDown(t)

	reply1 := &pb.VoteReply{
		Granted:proto.Bool(false),
		Term:proto.Uint64(1)}
	tt.mockExchange.EXPECT().Vote(gomock.Any(), gomock.Any()).Return(reply1, nil).Times(2)

	reply2 := &pb.VoteReply{
		Granted:proto.Bool(true),
		Term:proto.Uint64(1)}
	tt.mockExchange.EXPECT().Vote(gomock.Any(), gomock.Any()).Return(reply2, nil).Times(1)

	tt.candidate.Run()
	t.Log(tt.nodeMaster.store.CurrentTerm())
	if tt.nodeMaster.state != LEADER {
		t.Fail()
	}
}

func TestSingleReplica(t *testing.T) {
	tt := &CandidateComponentTest{}
	tt.setUp(t, peers2)
	defer tt.tearDown(t)

	tt.candidate.Run()
	if tt.nodeMaster.state != LEADER {
		t.Fail()
	}
}