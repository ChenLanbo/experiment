package raft
import (
	"testing"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"errors"
	"github.com/golang/protobuf/proto"
	"net"
	"syscall"
	"log"
)

type CandidateComponentTest struct {
	nodeMaster *NodeMaster
	candidate *Candidate

	s1 *grpc.Server
	s2 *grpc.Server
}

func (test *CandidateComponentTest) setUp() {
	peers := []string{"localhost:30001", "localhost:30002", "localhost:30003"}
	test.nodeMaster = NewNodeMaster(peers, 0)
	test.candidate = NewCandidate(test.nodeMaster)

	l1, err1 := net.Listen("tcp", peers[1])
	if err1 != nil {
		log.Fatal("Failed to opend socket.")
		syscall.Exit(1)
	}

	test.s1 = grpc.NewServer()
	pb.RegisterRaftServerServer(test.s1, &CandidateComponentTestRpcServer{})
	go func() {
		test.s1.Serve(l1)
	} ()

	l2, err2 := net.Listen("tcp", peers[2])
	if err2 != nil {
		log.Fatal("Failed to opend socket.")
		syscall.Exit(1)
	}

	test.s2 = grpc.NewServer()
	pb.RegisterRaftServerServer(test.s2, &CandidateComponentTestRpcServer{})
	go func() {
		test.s2.Serve(l2)
	} ()
}

func (test *CandidateComponentTest) tearDown() {
	test.nodeMaster.Stop()
	test.s1.Stop()
	test.s2.Stop()

	test.candidate = nil
	test.nodeMaster = nil
}

type CandidateComponentTestRpcServer struct {}

func (s *CandidateComponentTestRpcServer) Vote(ctx context.Context, request *pb.VoteRequest) (*pb.VoteReply, error) {
	reply := &pb.VoteReply{}
	reply.Granted = proto.Bool(true)
	reply.Term = proto.Uint64(0)
	return reply, nil
}

func (s *CandidateComponentTestRpcServer) Append(ctx context.Context, request *pb.AppendRequest) (*pb.AppendReply, error) {
	return nil, errors.New("Invalid operation")
}

func (s *CandidateComponentTestRpcServer) Put(ctx context.Context, request *pb.PutRequest) (*pb.PutReply, error) {
	return nil, errors.New("Invalid operation")
}

func TestVoteSelfOnce(t *testing.T) {
	test := &CandidateComponentTest{}
	test.setUp()
	defer test.tearDown()
	test.nodeMaster.state = CANDIDATE

	test.candidate.VoteSelfOnce()
	if test.nodeMaster.state != LEADER {
		t.Fatal("Should have become a leader")
		t.Fail()
	}
}

func TestCandidateProcessRequest(t *testing.T) {
	// TODO:
}