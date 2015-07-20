package raft
import (
	"testing"
	"golang.org/x/net/context"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"errors"
	"github.com/golang/protobuf/proto"
	"net"
	"google.golang.org/grpc"
	"syscall"
	"log"
)

type LeaderComponentTest struct {
	nodeMaster *NodeMaster
	leader *Leader

	s1 *grpc.Server
	s2 *grpc.Server
}

func (test *LeaderComponentTest) setUp() {
	peers := []string{"localhost:30001", "localhost:30002", "localhost:30003"}
	test.nodeMaster = NewNodeMaster(peers, 0)
	test.leader = NewLeader(test.nodeMaster)

	test.nodeMaster.store.IncrementCurrentTerm()

	// Setup rpc stub
	replicator := test.leader.logReplicators[0]
	l1, err1 := net.Listen("tcp", replicator.peer)
	if err1 != nil {
		log.Fatal("Failed to opend socket.")
		syscall.Exit(1)
	}

	test.s1 = grpc.NewServer()
	pb.RegisterRaftServerServer(test.s1, &testRpcServer{})
	go func() {
		test.s1.Serve(l1)
	} ()

	replicator = test.leader.logReplicators[1]
	l2, err2 := net.Listen("tcp", replicator.peer)
	if err2 != nil {
		log.Fatal("Failed to opend socket.")
		syscall.Exit(1)
	}

	test.s2 = grpc.NewServer()
	pb.RegisterRaftServerServer(test.s2, &testRpcServer{})
	go func() {
		test.s2.Serve(l2)
	} ()
}

func (test *LeaderComponentTest) tearDown() {
	test.nodeMaster.Stop()
	test.s1.Stop()
	test.s2.Stop()

	test.leader = nil
	test.nodeMaster = nil
}

type testRpcServer struct {}

func (s *testRpcServer) Vote(ctx context.Context, request *pb.VoteRequest) (*pb.VoteReply, error) {
	return nil, errors.New("Invalid operation")
}

func (s *testRpcServer) Append(ctx context.Context, request *pb.AppendRequest) (*pb.AppendReply, error) {
	reply := &pb.AppendReply{}
	reply.Success = proto.Bool(true)
	reply.Term = proto.Uint64(1)
	return reply, nil
}

func (s *testRpcServer) Put(ctx context.Context, request *pb.PutRequest) (*pb.PutReply, error) {
	return nil, errors.New("Invalid operation")
}

func TestLogReplicator(t *testing.T) {
	test := &LeaderComponentTest{}
	test.setUp()
	defer test.tearDown()

	replicator1 := test.leader.logReplicators[0]
	replicator1.ReplicateOnce()
	if replicator1.replicateIndex != test.nodeMaster.store.CommitIndex() {
		t.Fatal("ReplicateIndex should not be advancecd", replicator1.replicateIndex)
		t.Fail()
	}

	n := 8
	fakeData := make([]byte, 0)
	for i := 0; i < n; i++ {
		test.nodeMaster.store.Write(fakeData)
	}

	for i := 0; i < n; i++ {
		replicator1.ReplicateOnce()
	}
	if replicator1.replicateIndex != uint64(n) {
		t.Log("ReplicateIndex:", replicator1.replicateIndex, "\n")
		t.Fail()
	}
}

func TestLogCommitter(t *testing.T) {
	test := &LeaderComponentTest{}
	test.setUp()
	defer test.tearDown()

	n := 8
	fakeData := make([]byte, 0)
	for i := 0; i < n; i++ {
		test.nodeMaster.store.Write(fakeData)
		for _, replicator := range (test.leader.logReplicators) {
			replicator.ReplicateOnce()
		}
	}

	test.leader.logCommitter.CommitOnce()
	commitIndex := test.leader.nodeMaster.store.CommitIndex()
	if commitIndex != uint64(n) {
		t.Fatal("CommitIndex not advanced:", commitIndex, "expected:", n)
		t.Fail()
	}

	n++
	test.nodeMaster.store.Write(fakeData)

	test.leader.logCommitter.CommitOnce()
	commitIndex = test.leader.nodeMaster.store.CommitIndex()
	if commitIndex == uint64(n) {
		t.Fatal("CommitIndex should not be advanced:", commitIndex)
		t.Fail()
	}

	test.leader.logReplicators[0].ReplicateOnce()
	test.leader.logCommitter.CommitOnce()
	commitIndex = test.leader.nodeMaster.store.CommitIndex()
	if commitIndex != uint64(n) {
		t.Fatal("CommitIndex not advanced:", commitIndex, "expected:", n)
		t.Fail()
	}
}
