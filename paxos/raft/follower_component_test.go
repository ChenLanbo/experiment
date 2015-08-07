package raft
import (
	"testing"

	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"github.com/golang/mock/gomock"
	"github.com/chenlanbo/experiment/paxos/raft/test"
)

type FollowerComponentTest struct {
	mockCtrl *gomock.Controller
	mockExchange *test.MockMessageExchange

	peers []string
	me int

	nodeMaster *NodeMaster
	follower   *Follower
}

func (tt *FollowerComponentTest) setUp(t *testing.T) {
	tt.mockCtrl = gomock.NewController(t)
	tt.mockExchange = test.NewMockMessageExchange(tt.mockCtrl)

	tt.peers = []string{"localhost:10001", "localhost:10002", "localhost:10003"}
	tt.me = 0

	tt.nodeMaster = NewNodeMaster(tt.mockExchange, tt.peers, tt.me)
	tt.follower = NewFollower(tt.nodeMaster)
}

func (tt *FollowerComponentTest) tearDown(t *testing.T) {
	tt.nodeMaster.Stop()
	tt.follower = nil
	tt.nodeMaster = nil

	tt.mockCtrl.Finish()
	tt.mockExchange = nil
	tt.mockCtrl = nil
}

func TestTimeoutWithNoRequests(t *testing.T) {
	tt := FollowerComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	tt.follower.ProcessOneRequest()

	if tt.nodeMaster.state != CANDIDATE {
		t.Fatal("State is not CANDIDATE but", tt.nodeMaster.state)
		t.Fail()
	}
}

func TestVote(t *testing.T) {
	tt := FollowerComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	voteRequest1 := &pb.VoteRequest{
		Term:proto.Uint64(1),
		CandidateId:proto.String(tt.peers[1]),
		LastLogTerm:proto.Uint64(0),
		LastLogIndex:proto.Uint64(0)}

	op := NewRaftOperation(NewRaftRequest(voteRequest1, nil))
	tt.nodeMaster.OpsQueue.Push(op)

	tt.follower.ProcessOneRequest()
	if tt.nodeMaster.state != FOLLOWER {
		t.Fail()
	}
	if tt.nodeMaster.store.CurrentTerm() != 1 {
		t.Fail()
	}

	reply := <- op.Callback
	if reply.VoteReply == nil {
		t.Fatal("Empty vote reply")
		t.Fail()
	}
	if !*reply.VoteReply.Granted {
		t.Fatal("Vote not granted")
		t.Fail()
	}
	close(op.Callback)

	// Another candidate send vote in the same term
	voteRequest2 := &pb.VoteRequest{
		Term:proto.Uint64(1),
		CandidateId:proto.String(tt.peers[2]),
		LastLogTerm:proto.Uint64(0),
		LastLogIndex:proto.Uint64(0)}

	op = NewRaftOperation(NewRaftRequest(voteRequest2, nil))
	tt.nodeMaster.OpsQueue.Push(op)

	tt.follower.ProcessOneRequest()
	if tt.nodeMaster.state != FOLLOWER {
		t.Fail()
	}

	reply = <- op.Callback
	if reply.VoteReply == nil {
		t.Fatal("Empty vote reply")
		t.Fail()
	}
	if *reply.VoteReply.Granted {
		t.Fatal("Vote should not be granted")
		t.Fail()
	}
	close(op.Callback)
}

func TestAppend(t *testing.T) {
	tt := FollowerComponentTest{}
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

	op := NewRaftOperation(NewRaftRequest(nil, appendRequest))
	tt.nodeMaster.OpsQueue.Push(op)

	tt.follower.ProcessOneRequest()
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
	if !*reply.AppendReply.Success {
		t.Fatal("Append failed")
		t.Fail()
	}
}

func TestVoteAndAppend(t *testing.T) {
	// TODO(lanbochen): fill this test
}