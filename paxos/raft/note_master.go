package raft
import (
	"github.com/chenlanbo/experiment/paxos/raft/store"
	"github.com/chenlanbo/experiment/paxos/raft/rpc"
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
		for {
			switch nodeMaster.state {
			case LEADER:
				break
			case FOLLOWER:
				break
			case CANDIDATE:
				break
			}
		}
	} ()
}

func (nodeMaster *NodeMaster) Stop() {
	nodeMaster.OpsQueue.Close()
}

func (nodeMaster *NodeMaster) runLeader() {

}

func (nodeMaster *NodeMaster) runFollower() {

}

func (NodeMaster *NodeMaster) runCandidate() {

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

	return nodeMaster
}
