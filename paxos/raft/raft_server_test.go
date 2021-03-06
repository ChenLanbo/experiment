package raft
import (
	"time"
	"testing"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"github.com/chenlanbo/experiment/paxos/raft/rpc"
	"errors"
	"math/rand"
	"reflect"
	"strconv"
)

var (
	peers1 = []string{"localhost:30001", "localhost:30002", "localhost:30003"}
 	peers2 = []string{"localhost:30001"}
	peers3 = []string{"localhost:30001", "localhost:30002", "localhost:30003", "localhost:30004", "localhost:30005"}
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
	time.Sleep(time.Second * 3)

	if sendPutToReplicas(peers2, "abc", []byte("abc")) != nil {
		t.Error("Put not succeeded")
	}
	if sendGetToReplicas(peers2, "abc", []byte("abc")) != nil {
		t.Error("Get not succeeded")
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

	if sendPutToReplicas(peers1, "abc", []byte("abc")) != nil {
		t.Fail()
	}
	
	// Check all replicas get the put
	for _, raft := range tt.servers {
		if raft.nodeMaster.store.LatestIndex() < 1 {
			t.Fail()
		}
	}
}

func TestRaftServerThreeReplicaWithOneDown(t *testing.T) {
	tt := &RaftServerTest{}
	tt.setUp(t, peers1)
	defer tt.tearDown(t)

	// The third node doesn't start
	tt.servers[0].Start()
	tt.servers[1].Start()

	time.Sleep(time.Second * 5)

	if sendPutToReplicas(peers1, "abc", []byte("abc")) != nil {
		t.Fail()
	}

	if tt.servers[0].nodeMaster.store.LatestIndex() < 1 {
		t.Fail()
	}
	if tt.servers[0].nodeMaster.store.LatestIndex() < 1 {
		t.Fail()
	}

	// The third node comes back again
	tt.servers[2].Start()
	time.Sleep(time.Second * 5)

	if tt.servers[2].nodeMaster.store.LatestIndex() < 1 {
		t.Fail()
	}
}

func TestRaftServerThreeReplicaWithLeaderFailOver(t *testing.T)  {
	tt := &RaftServerTest{}
	tt.setUp(t, peers1)
	defer tt.tearDown(t)

	for _, raft := range tt.servers {
		raft.Start()
	}
	time.Sleep(time.Second * 5)

	if sendPutToReplicas(peers1, "abc", []byte("abc")) != nil {
		t.Fail()
	}

	// Stop the leader
	curLeader := -1
	for id, raft := range tt.servers {
		if raft.nodeMaster.state == LEADER {
			curLeader = id
			raft.Stop()
		}
	}
	if curLeader == -1 {
		t.Error("No leader elected")
	}

	// A new leader should be elected and put should succeed
	time.Sleep(time.Second * 6)
	tempPeers := make([]string, 0)
	for id, _ := range peers1 {
		if id != curLeader {
			tempPeers = append(tempPeers, peers1[id])
		}
	}
	if sendPutToReplicas(tempPeers, "abc", []byte("abc")) != nil {
		t.Error("Put should have succeeded with remaining peers")
	}

	for id, raft := range tt.servers {
		if id == curLeader {
			continue
		}

		if raft.nodeMaster.store.LatestIndex() != 2 {
			t.Fail()
		}
	}
}

func TestRaftServerThreeReplicaWithLogNotOnMajority(t *testing.T) {
	tt := &RaftServerTest{}
	tt.setUp(t, peers1)
	defer tt.tearDown(t)

	for id, raft := range tt.servers {
		raft.nodeMaster.store.IncrementCurrentTerm()
		raft.nodeMaster.store.WriteKeyValue("abc", []byte("abc"))
		if id == 2 {
			// Peer 2's log not on majority
			raft.nodeMaster.store.WriteKeyValue("abc1", []byte("abc1"))
		}
	}

	tt.servers[0].Start()
	tt.servers[1].Start()
	time.Sleep(time.Second * 3)

	// Peer 2's log will be overridden
	tt.servers[2].Start()
	time.Sleep(time.Second * 3)

	if sendPutToReplicas(peers1, "abc", []byte("abc")) != nil {
		t.Fail()
	}
	time.Sleep(time.Second)

	l := tt.servers[2].nodeMaster.store.Read(2)
	if l.GetData() == nil {
		t.Fail()
	}
	if l.GetData().GetKey() != "abc" {
		t.Fail()
	}
}

func TestRaftServerThreeReplicaNotUpToDatePeerNotLeader(t *testing.T) {
	tt := &RaftServerTest{}
	tt.setUp(t, peers1)
	defer tt.tearDown(t)

	for id, raft := range tt.servers {
		raft.nodeMaster.store.IncrementCurrentTerm()
		raft.nodeMaster.store.WriteKeyValue("abc", []byte("abc"))
		if id != 2 {
			// Peer 2's log not on majority
			raft.nodeMaster.store.WriteKeyValue("abc1", []byte("abc1"))
		}
	}

	tt.servers[2].Start()
	time.Sleep(time.Second * 1)

	tt.servers[0].Start()
	tt.servers[1].Start()
	time.Sleep(time.Second * 5)

	for id, raft := range tt.servers {
		if id == 2 && raft.nodeMaster.state == LEADER {
			t.Fail()
		}
	}

	if sendPutToReplicas(peers1, "abc", []byte("abc")) != nil {
		t.Fail()
	}
}

func TestRaftServerThreeReplicaMajorityDown(t *testing.T) {
	tt := &RaftServerTest{}
	tt.setUp(t, peers1)
	defer tt.tearDown(t)

	tt.servers[0].Start()
	time.Sleep(time.Second * 2)

	if tt.servers[0].nodeMaster.state == LEADER {
		t.Fail()
	}
	if sendPutToReplicas(peers1, "abc", []byte("abc")) == nil {
		t.Fail()
	}

	// Majority come back
	tt.servers[1].Start()
	tt.servers[2].Start()
	time.Sleep(time.Second * 3)
	if sendPutToReplicas(peers1, "def", []byte("def")) != nil {
		t.Error("Majority should have come back")
	}
}

func TestRaftServerFiveReplica(t *testing.T) {
	tt := &RaftServerTest{}
	tt.setUp(t, peers3)
	defer tt.tearDown(t)

	for _, raft := range tt.servers {
		raft.Start()
	}
	time.Sleep(time.Second * 5)

	for i := 0; i < 100; i++ {
		if sendPutToReplicas(peers3, strconv.Itoa(i), []byte(strconv.Itoa(i))) != nil {
			t.Fail()
		}
	}
	
	for i := 0; i < 100; i++ {
		if sendGetToReplicas(peers3, strconv.Itoa(i), []byte(strconv.Itoa(i))) != nil {
			t.Fail()
		}
	}
}

// Util function
func sendPutToReplicas(peers []string, key string, value []byte) error {
	exchange := rpc.NewMessageExchange()
	request := &pb.PutRequest{
		Key:proto.String(key), Value:[]byte(value)}
	tries := 0
	to := peers[0]

	for ; tries < 5; tries++ {
		reply, err := exchange.Put(to, request)
		if err != nil {
			continue
		}
		if reply.GetSuccess() {
			break
		} else {
			to = reply.GetLeaderId()
			if to == "" {
				to = peers[rand.Intn(len(peers))]
			}
		}
	}

	if tries == 5 {
		return errors.New("Fail to put")
	}

	return nil
}

func sendGetToReplicas(peers[]string, key string, expectedValue []byte) error {
	exchange := rpc.NewMessageExchange()
	request := &pb.GetRequest{
		Key:proto.String(key)}
	tries := 0
	to := peers[0]

	for ; tries < 5; tries++ {
		reply, err := exchange.Get(to, request)
		if err != nil {
			continue
		}
		if reply.GetSuccess() {
			if !reflect.DeepEqual(expectedValue, reply.GetValue()) {
				return errors.New("Got bad value")
			}
			break
		} else {
			to = reply.GetLeaderId()
			if to == "" {
				to = peers[rand.Intn(len(peers))]
			}
		}
	}
	return nil
}