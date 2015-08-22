package raft
import (
	"github.com/chenlanbo/experiment/paxos/raft/store"
	"github.com/chenlanbo/experiment/paxos/raft/rpc"
	"log"
	"sync/atomic"
)

type NodeState int
const (
	LEADER NodeState = iota + 1
	FOLLOWER
	CANDIDATE
)

type NodeMaster struct {
	OpsQueue *OperationQueue
	Exchange rpc.MessageExchange

	state NodeState
	votedLeader string

	store *store.Store
	inflightRequests map[uint64]*RaftOperation

	peers []string
	me    int

	stopped int32
	stopChan chan bool
}

func (nodeMaster *NodeMaster) MyEndpoint() string {
	return nodeMaster.peers[nodeMaster.me]
}

func (nodeMaster *NodeMaster) PeerEndpoints() []string {
	endpoints := make([]string, 0)
	for id, peer := range(nodeMaster.peers) {
		if id == nodeMaster.me {
			continue
		}
		endpoints = append(endpoints, peer)
	}
	return endpoints
}

func (nodeMaster *NodeMaster) Start() {
	go func() {
		for !nodeMaster.Stopped() {
			switch nodeMaster.state {
			case LEADER:
				log.Println(nodeMaster.MyEndpoint(), "run as a leader")
				nodeMaster.runLeader()
			case FOLLOWER:
				log.Println(nodeMaster.MyEndpoint(), "run as a follower")
				nodeMaster.runFollower()
			case CANDIDATE:
				log.Println(nodeMaster.MyEndpoint(), "run as a candidate")
				nodeMaster.runCandidate()
			}
		}
	} ()
}

func (nodeMaster *NodeMaster) Stop() {
	atomic.StoreInt32(&nodeMaster.stopped, 1)
	nodeMaster.OpsQueue.Close()
	nodeMaster.stopChan <- true
}

func (nodeMaster *NodeMaster) Stopped() bool {
	return atomic.LoadInt32(&nodeMaster.stopped) == 1
}

func (nodeMaster *NodeMaster) runLeader() {
	leader := NewLeader(nodeMaster)
	notify := make(chan bool, 1)

	go func () {
		leader.Run()
		log.Println(nodeMaster.MyEndpoint(), "stop as a leader")
		notify <- true
	} ()

	select {
	case <-notify:
		break
	case <-nodeMaster.stopChan:
		break
	}
	leader.Stop()
	close(notify)
}

func (nodeMaster *NodeMaster) runFollower() {
	follower := NewFollower(nodeMaster)
	notify := make(chan bool, 1)

	go func () {
		follower.Run()
		log.Println(nodeMaster.MyEndpoint(), "stop as a follower")
		notify <- true
	} ()

	select {
	case <-notify:
		break
	case <-nodeMaster.stopChan:
		break
	}
	follower.Stop()
	close(notify)
}

func (nodeMaster *NodeMaster) runCandidate() {
	candidate := NewCandidate(nodeMaster)
	notify := make(chan bool, 1)

	go func () {
		candidate.Run()
		log.Println(nodeMaster.MyEndpoint(), "stop as a candidate")
		notify <- true
	} ()

	select {
	case <-notify:
		break
	case <-nodeMaster.stopChan:
		break
	}
	candidate.Stop()
	close(notify)
}

func NewNodeMaster(exchange rpc.MessageExchange, peers []string, me int) (*NodeMaster) {
	nodeMaster := &NodeMaster{}

	nodeMaster.OpsQueue = NewOperationQueue()
	nodeMaster.inflightRequests = make(map[uint64]*RaftOperation)
	nodeMaster.Exchange = exchange
	nodeMaster.state = FOLLOWER
	nodeMaster.store = store.NewStore()
	nodeMaster.votedLeader = ""

	nodeMaster.peers = make([]string, len(peers))
	copy(nodeMaster.peers, peers)
	nodeMaster.me = me

	nodeMaster.stopped = 0
	nodeMaster.stopChan = make(chan bool, 4)

	return nodeMaster
}
