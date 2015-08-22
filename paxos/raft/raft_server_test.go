package raft
import (
	"time"
	"testing"
	"google.golang.org/grpc"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"golang.org/x/net/context"
	"github.com/golang/protobuf/proto"
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

	c, err := grpc.Dial(peers2[0])
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	defer c.Close()

	cli := pb.NewRaftClient(c)
	reply, err1 := cli.Put(context.Background(), &pb.PutRequest{
		Key:proto.String("abc"), Value:[]byte("abc")})
	if err1 != nil {
		t.Logf("Server returns error %s", err1)
		t.Fail()
	}

	if !reply.GetSuccess() {
		t.Logf("Put not succeeded")
		t.Fail()
	}
}