package raft
import (
	"testing"

	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
)

var (
	nodeMaster *NodeMaster = nil
	follower   *Follower = nil

	candidateId string
	leaderId    string
)

func setUpFollower() {
	peers := make([]string, 3)
	peers[0], peers[1], peers[2] = "localhost:10001", "localhost:10002", "localhost:10003"
	candidateId, leaderId = peers[1], peers[1]

	nodeMaster = NewNodeMaster(peers, 0)
	follower = NewFollower(nodeMaster)
}

func tearDownFollower() {
	nodeMaster.Stop()

	follower = nil
	nodeMaster = nil
}

func TestTimeoutWithNoRequests(t *testing.T) {
	setUpFollower()
	defer tearDownFollower()

	follower.ProcessOneRequest()

	if nodeMaster.state != CANDIDATE {
		t.Fatal("State is not CANDIDATE but", nodeMaster.state)
		t.Fail()
	}
}

func TestVote(t *testing.T) {
	setUpFollower()
	defer tearDownFollower()

	voteRequest1 := &pb.VoteRequest{
		Term:proto.Uint64(1),
		CandidateId:proto.String(candidateId),
		LastLogTerm:proto.Uint64(0),
		LastLogIndex:proto.Uint64(0)}

	op := NewRaftOperation(NewRaftRequest(voteRequest1, nil))
	nodeMaster.OpsQueue.Push(op)

	follower.ProcessOneRequest()
	if nodeMaster.state != FOLLOWER {
		t.Fail()
	}
	if nodeMaster.store.CurrentTerm() != 1 {
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
		CandidateId:proto.String("localhost:10003"),
		LastLogTerm:proto.Uint64(0),
		LastLogIndex:proto.Uint64(0)}

	op = NewRaftOperation(NewRaftRequest(voteRequest2, nil))
	nodeMaster.OpsQueue.Push(op)

	follower.ProcessOneRequest()
	if nodeMaster.state != FOLLOWER {
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
	setUpFollower()
	defer tearDownFollower()

	leaderId := "localhost:10002"

	appendRequest := &pb.AppendRequest{
		Term:proto.Uint64(1),
		LeaderId:proto.String(leaderId),
		PrevLogIndex:proto.Uint64(0),
		PrevLogTerm:proto.Uint64(0),
		CommitIndex:proto.Uint64(0),
		Logs:[]*pb.Log{
			&pb.Log{Term:proto.Uint64(1), LogId:proto.Uint64(1)},
			&pb.Log{Term:proto.Uint64(1), LogId:proto.Uint64(2)}}}

	op := NewRaftOperation(NewRaftRequest(nil, appendRequest))
	nodeMaster.OpsQueue.Push(op)

	follower.ProcessOneRequest()
	if nodeMaster.store.CurrentTerm() != 1 {
		t.Fatal("Follower's term not advanced")
		t.Fail()
	}
	if nodeMaster.store.LatestIndex() != 2 {
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

}