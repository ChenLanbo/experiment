package raft
import (
	"time"
	"testing"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"github.com/chenlanbo/experiment/paxos/raft/rpc"
	"errors"
)

var (
	peers1 = []string{"localhost:30001", "localhost:30002", "localhost:30003"}
 	peers2 = []string{"localhost:30001"}
)

type RaftServerTest struct {
	peers []string
	me int

	servers []*RaftServer
}

func (tt *RaftServerTest) setUp(t *testing.T, peers []string) {
	tt.peers = peers
	tt.servers = make([]*RaftServer, len(peers))
	tt.me = 0
	for id, _ := range tt.peers {
		tt.servers[id] = NewRaftServer(tt.peers, id)
	}
}

func (tt *RaftServerTest) tearDown(t *testing.T) {
	for _, raft := range tt.servers {
		raft.Stop()
	}
}

func TestRaftServerOneReplica(t *testing.T) {
	tt := &RaftServerTest{}
	tt.setUp(t, peers2)
	defer tt.tearDown(t)

	for _, raft := range tt.servers {
		raft.Start()
	}
	time.Sleep(time.Second * 5)

	exchange := rpc.NewMessageExchange()
	request := &pb.PutRequest{
		Key:proto.String("abc"), Value:[]byte("abc")}

	reply, _ := exchange.Put(peers2[0], request)
	if !reply.GetSuccess() {
		t.Logf("Put not succeeded")
		t.Fail()
	}
}

func TestRaftServerThreeReplica(t *testing.T) {
	tt := &RaftServerTest{}
	tt.setUp(t, peers1)
	defer tt.tearDown(t)

	for _, raft := range tt.servers {
		raft.Start()
	}

	time.Sleep(time.Second * 5)

	err := sendPutToReplicas(peers1)
	if err != nil {
		t.Fail()
	}
	
	// Check all replicas get the put
	for _, raft := range tt.servers {
		if raft.nodeMaster.store.LatestIndex() < 1 {
			t.Fail()
		}
	}
}

func sendPutToReplicas(peers []string) error {
	exchange := rpc.NewMessageExchange()
	request := &pb.PutRequest{
		Key:proto.String("abc"), Value:[]byte("abc")}
	tries := 0
	to := peers[0]

	for ; tries < 5; tries++ {
		reply, _ := exchange.Put(to, request)
		if reply.GetSuccess() {
			break
		} else {
			to = reply.GetLeaderId()
		}
	}

	if tries == 5 {
		return errors.New("Fail to put")
	}

	return nil
}