package raft
import (
	"testing"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"github.com/golang/mock/gomock"
	"github.com/chenlanbo/experiment/paxos/raft/test"
)

type LeaderComponentTest struct {
	mockCtrl *gomock.Controller
	mockExchange *test.MockMessageExchange

	nodeMaster *NodeMaster
	leader *Leader

	peers []string
	me int
}

func (tt *LeaderComponentTest) setUp(t *testing.T) {
	tt.mockCtrl = gomock.NewController(t)
	tt.mockExchange = test.NewMockMessageExchange(tt.mockCtrl)

	tt.peers = []string{"localhost:30001", "localhost:30002", "localhost:30003"}
	tt.me = 0

	tt.nodeMaster = NewNodeMaster(tt.mockExchange, tt.peers, tt.me)
	tt.leader = NewLeader(tt.nodeMaster)

	tt.nodeMaster.state = LEADER
	tt.nodeMaster.store.IncrementCurrentTerm()
}

func (tt *LeaderComponentTest) tearDown(t *testing.T) {
	tt.nodeMaster.Stop()
	tt.leader = nil
	tt.nodeMaster = nil

	tt.mockCtrl.Finish()
	tt.mockExchange = nil
	tt.mockCtrl = nil
}

func TestLogReplicatorReplicateIndexAdvanced(t *testing.T) {
	tt := &LeaderComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	n := 8
	fakeData := make([]byte, 0)
	for i := 0; i < n; i++ {
		tt.nodeMaster.store.Write(fakeData)
	}

	reply := &pb.AppendReply{
		Success:proto.Bool(true),
		Term:proto.Uint64(1)}
	tt.mockExchange.EXPECT().Append(gomock.Any(), gomock.Any()).Return(reply, nil).AnyTimes()

	replicator := tt.leader.replicators[0]
	for i := 0; i < n; i++ {
		replicator.ReplicateOnce()
	}

	if replicator.replicateIndex != uint64(n) {
		t.Log("ReplicateIndex:", replicator.replicateIndex, "\n")
		t.Fail()
	}
	replicator.Stop()
}

func TestLogReplicatorReplicateFromPreviousLog(t *testing.T) {
	tt := &LeaderComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	replicator := tt.leader.replicators[0]

	n := 8
	fakeData := make([]byte, 0)
	for i := 0; i < n; i++ {
		tt.nodeMaster.store.Write(fakeData)
	}
	tt.nodeMaster.store.SetCommitIndex(uint64(n))
	replicator.replicateIndex = uint64(n)

	reply1 := &pb.AppendReply{
		Success:proto.Bool(false),
		Term:proto.Uint64(tt.nodeMaster.store.CurrentTerm())}
	tt.mockExchange.EXPECT().Append(gomock.Any(), gomock.Any()).Return(reply1, nil)
	replicator.ReplicateOnce()

	// ReplicateIndex moves backward
	if replicator.replicateIndex != uint64(n - 1) {
		t.Fatal("ReplicateIndex should move backward")
		t.Fail()
	}

	tt.nodeMaster.store.IncrementCurrentTerm()
	tt.nodeMaster.store.Write(fakeData)
	n++

	reply2 := &pb.AppendReply{
		Success:proto.Bool(true),
		Term:proto.Uint64(tt.nodeMaster.store.CurrentTerm())}
	tt.mockExchange.EXPECT().Append(gomock.Any(), gomock.Any()).Return(reply2, nil)
	tt.mockExchange.EXPECT().Append(gomock.Any(), gomock.Any()).Return(reply2, nil)
	replicator.ReplicateOnce()
	replicator.ReplicateOnce()

	// ReplicateIndex moves forward
	if replicator.replicateIndex != uint64(n) {
		t.Fatal("ReplicateIndex should move forward")
		t.Fail()
	}
}

func TestLogCommitter(t *testing.T) {
	tt := &LeaderComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	reply := &pb.AppendReply{
		Success:proto.Bool(true),
		Term:proto.Uint64(1)}
	tt.mockExchange.EXPECT().Append(gomock.Any(), gomock.Any()).Return(reply, nil).AnyTimes()

	fakeData := make([]byte, 0)
	n := 8
	for i := 0; i < n; i++ {
		tt.nodeMaster.store.Write(fakeData)
		for _, replicator := range tt.leader.replicators {
			replicator.ReplicateOnce()
		}
	}

	tt.leader.committer.CommitOnce()
	commitIndex := tt.leader.nodeMaster.store.CommitIndex()
	if commitIndex != uint64(n) {
		t.Fatal("CommitIndex not advanced:", commitIndex, "expected:", n)
		t.Fail()
	}

	n++
	tt.nodeMaster.store.Write(fakeData)
	tt.leader.committer.CommitOnce()
	commitIndex = tt.leader.nodeMaster.store.CommitIndex()
	if commitIndex == uint64(n) {
		t.Fatal("CommitIndex should not be advanced:", commitIndex)
		t.Fail()
	}

	tt.leader.replicators[0].ReplicateOnce()
	tt.leader.committer.CommitOnce()
	commitIndex = tt.leader.nodeMaster.store.CommitIndex()
	if commitIndex != uint64(n) {
		t.Fatal("CommitIndex not advanced:", commitIndex, "expected:", n)
		t.Fail()
	}
}
