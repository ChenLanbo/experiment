package raft
import (
	"testing"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"github.com/golang/mock/gomock"
	"github.com/chenlanbo/experiment/paxos/raft/test"
	"time"
	"reflect"
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

func TestLeaderLogReplicatorReplicateIndexAdvanced(t *testing.T) {
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

func TestLeaderLogReplicatorReplicateFromPreviousLog(t *testing.T) {
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

func TestLeaderLogCommitterCommit(t *testing.T) {
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

func TestLeaderProcessorHandlePut(t *testing.T) {
	tt := &LeaderComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	appendReply := &pb.AppendReply{
		Success:proto.Bool(true),
		Term:proto.Uint64(1)}
	tt.mockExchange.EXPECT().Append(gomock.Any(), gomock.Any()).Return(appendReply, nil).AnyTimes()

	request := &pb.PutRequest{
		Key:proto.String("abc"),
		Value:[]byte("abc")}
	op := NewRaftOperation(WithPutRequest(EmptyRaftRequest(), request))
	tt.nodeMaster.OpsQueue.Push(op)

	tt.leader.processor.ProcessOnce()
	tt.leader.replicators[0].ReplicateOnce()
	tt.leader.committer.CommitOnce()

	reply := <- op.Callback
	if reply.PutReply == nil {
		t.Fail()
	}
	if !reply.PutReply.GetSuccess() {
		t.Fail()
	}
}

func TestLeaderProcessorHandleVote(t *testing.T) {
	tt := &LeaderComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	request1 := &pb.VoteRequest{
		Term:proto.Uint64(1),
		CandidateId:proto.String(tt.peers[1]),
		LastLogTerm:proto.Uint64(0),
		LastLogIndex:proto.Uint64(0)}
	op1 := NewRaftOperation(WithVoteRequest(EmptyRaftRequest(), request1))
	tt.nodeMaster.OpsQueue.Push(op1)

	if !tt.leader.processor.ProcessOnce() {
		t.Fail()
	}

	reply1 := <- op1.Callback
	if reply1.VoteReply == nil {
		t.Fail()
	}

	request2 := &pb.VoteRequest{
		Term:proto.Uint64(2),
		CandidateId:proto.String(tt.peers[1]),
		LastLogTerm:proto.Uint64(0),
		LastLogIndex:proto.Uint64(0)}
	op2 := NewRaftOperation(WithVoteRequest(EmptyRaftRequest(), request2))
	tt.nodeMaster.OpsQueue.Push(op2)

	if tt.leader.processor.ProcessOnce() {
		t.Fail()
	}

	reply2 := <- op2.Callback
	if reply2.VoteReply == nil {
		t.Fail()
	}

	newTerm := <- tt.leader.newTermChan
	if newTerm != 2 {
		t.Fail()
	}
}

func TestLeaderProcessorHandleGet(t *testing.T) {
	tt := &LeaderComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	appendReply := &pb.AppendReply{
		Success:proto.Bool(true),
		Term:proto.Uint64(1)}
	tt.mockExchange.EXPECT().Append(gomock.Any(), gomock.Any()).Return(appendReply, nil).AnyTimes()

	putRequest := &pb.PutRequest{
		Key:proto.String("abc"),
		Value:[]byte("abc")}
	op1 := NewRaftOperation(WithPutRequest(EmptyRaftRequest(), putRequest))
	defer close(op1.Callback)
	tt.nodeMaster.OpsQueue.Push(op1)

	tt.leader.processor.ProcessOnce()
	tt.leader.replicators[0].ReplicateOnce()
	tt.leader.committer.CommitOnce()

	reply := <- op1.Callback
	if reply.PutReply == nil || !reply.PutReply.GetSuccess() {
		t.Fail()
	}

	getRequest := &pb.GetRequest{
		Key:proto.String("abc")}
	op2 := NewRaftOperation(WithGetRequest(EmptyRaftRequest(), getRequest))
	defer close(op2.Callback)
	tt.nodeMaster.OpsQueue.Push(op2)

	tt.leader.processor.ProcessOnce()

	reply = <- op2.Callback
	if reply.GetReply == nil || !reply.GetReply.GetSuccess() {
		t.Fail()
	}
	if !reflect.DeepEqual([]byte("abc"), reply.GetReply.GetValue()) {
		t.Fail()
	}
}

func TestLeaderRun(t *testing.T) {
	tt := &LeaderComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	appendReply := &pb.AppendReply{
		Success:proto.Bool(true),
		Term:proto.Uint64(1)}
	tt.mockExchange.EXPECT().Append(gomock.Any(), gomock.Any()).Return(appendReply, nil).AnyTimes()

	go func() {
		tt.leader.Run()
	} ()

	// Test put
	request := &pb.PutRequest{
		Key:proto.String("abc"),
		Value:[]byte("abc")}
	op := NewRaftOperation(WithPutRequest(EmptyRaftRequest(), request))
	tt.nodeMaster.OpsQueue.Push(op)

	reply := <- op.Callback
	if reply.PutReply == nil {
		t.Fail()
	}
	if !reply.PutReply.GetSuccess() {
		t.Fail()
	}

	// Test stop
	tt.leader.Stop()
	if !tt.leader.committer.Stopped() || !tt.leader.processor.Stopped() {
		t.Fail()
	}
	for _, replicator := range tt.leader.replicators {
		if !replicator.Stopped() {
			t.Fail()
		}
	}
}

func TestLeaderRun2(t *testing.T) {
	tt := &LeaderComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	appendReply := &pb.AppendReply{
		Success:proto.Bool(true),
		Term:proto.Uint64(1)}
	tt.mockExchange.EXPECT().Append(gomock.Any(), gomock.Any()).Return(appendReply, nil).AnyTimes()

	go func() {
		tt.leader.Run()
	} ()

	// Test vote at higher term
	request := &pb.VoteRequest{
		Term:proto.Uint64(2),
		CandidateId:proto.String(tt.peers[1]),
		LastLogTerm:proto.Uint64(0),
		LastLogIndex:proto.Uint64(0)}
	op := NewRaftOperation(WithVoteRequest(EmptyRaftRequest(), request))
	tt.nodeMaster.OpsQueue.Push(op)

	reply := <- op.Callback
	if reply.VoteReply == nil {
		t.Fail()
	}

	time.Sleep(time.Second)
	if !tt.leader.Stopped() {
		t.Fail()
	}
}

func TestLeaderRun3(t *testing.T) {
	tt := &LeaderComponentTest{}
	tt.setUp(t)
	defer tt.tearDown(t)

	// Leader gets evicted, put request should not success
	appendReply := &pb.AppendReply{
		Success:proto.Bool(false),
		Term:proto.Uint64(2)}
	tt.mockExchange.EXPECT().Append(gomock.Any(), gomock.Any()).Return(appendReply, nil).AnyTimes()

	request := &pb.PutRequest{
		Key:proto.String("abc"),
		Value:[]byte("abc")}
	op := NewRaftOperation(WithPutRequest(EmptyRaftRequest(), request))
	tt.nodeMaster.OpsQueue.Push(op)

	go func() {
		tt.leader.Run()
	} ()

	reply := <- op.Callback
	if reply.PutReply == nil {
		t.Fail()
	}
	if reply.PutReply.GetSuccess() {
		t.Fail()
	}

	time.Sleep(time.Second)
	if tt.nodeMaster.state != FOLLOWER {
		t.Fail()
	}
	if !tt.leader.Stopped() {
		t.Fail()
	}
}