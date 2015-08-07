package raft
import (
	"testing"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"github.com/golang/mock/gomock"
	"github.com/chenlanbo/experiment/paxos/raft/test"
)

type CandidateComponentTest struct {
	mockCtrl *gomock.Controller
	mockExchange *test.MockMessageExchange

	peers []string
	me int

	nodeMaster *NodeMaster
	candidate *Candidate
}

func (tt *CandidateComponentTest) setUp(t *testing.T) {
	tt.mockCtrl = gomock.NewController(t)
	tt.mockExchange = test.NewMockMessageExchange(tt.mockCtrl)

	tt.peers = []string{"localhost:30001", "localhost:30002", "localhost:30003"}
	tt.me = 0

	tt.nodeMaster = NewNodeMaster(tt.mockExchange, tt.peers, tt.me)
	tt.candidate = NewCandidate(tt.nodeMaster)
}

func (tt *CandidateComponentTest) tearDown(t *testing.T) {
	tt.nodeMaster.Stop()
	tt.candidate = nil
	tt.nodeMaster = nil

	tt.mockCtrl.Finish()
	tt.mockExchange = nil
	tt.mockCtrl = nil
}

func TestVoteBothGranted(t *testing.T) {
	tt := &CandidateComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)
	tt.nodeMaster.state = CANDIDATE

	reply := &pb.VoteReply{
		Granted:proto.Bool(true),
		Term:proto.Uint64(0)}
	tt.mockExchange.EXPECT().Vote(gomock.Any(), gomock.Any()).Return(reply, nil)

	tt.candidate.VoteSelfOnce()
	if tt.nodeMaster.state != LEADER {
		t.Fatal("Should have become a leader")
		t.Fail()
	}
}

func TestVoteOnlyOneGranted(t *testing.T) {
	tt := &CandidateComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)
	tt.nodeMaster.state = CANDIDATE

	reply1 := &pb.VoteReply{
		Granted:proto.Bool(true),
		Term:proto.Uint64(0)}
	tt.mockExchange.EXPECT().Vote(tt.peers[2], gomock.Any()).Return(reply1, nil)

	reply2 := &pb.VoteReply{
		Granted:proto.Bool(false),
		Term:proto.Uint64(1)}
	tt.mockExchange.EXPECT().Vote(tt.peers[1], gomock.Any()).Return(reply2, nil)

	tt.candidate.VoteSelfOnce()
	if tt.nodeMaster.state != LEADER {
		t.Fatal("Should have become a leader")
		t.Fail()
	}
}

func TestVoteHigherTermReturned(t *testing.T) {
	tt := &CandidateComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)
	tt.nodeMaster.state = CANDIDATE

	reply := &pb.VoteReply{
		Granted:proto.Bool(false),
		Term:proto.Uint64(2)}
	tt.mockExchange.EXPECT().Vote(gomock.Any(), gomock.Any()).Return(reply, nil)

	tt.candidate.VoteSelfOnce()
	if tt.nodeMaster.state != FOLLOWER {
		t.Fatal("Should have become a follower")
		t.Fail()
	}
}

func TestCandidateProcessRequest(t *testing.T) {
	// TODO(lanbochen): fill this test
}
