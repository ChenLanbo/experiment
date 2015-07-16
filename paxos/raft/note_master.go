package raft
import (
	"github.com/chenlanbo/experiment/paxos/raft/store"
)

type NodeState int
const (
	LEADER NodeState = iota + 1
	FOLLOWER
	CANDIDATE
)

type NodeMaster struct {

	OpsQueue *OperationQueue

	state NodeState
	votedLeader string

	store *store.Store

	peers []string
	me    int
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

func NewNodeMaster(peers []string, me int) (*NodeMaster) {
	nodeMaster := &NodeMaster{}

	nodeMaster.OpsQueue = NewOperationQueue()
	nodeMaster.state = FOLLOWER
	nodeMaster.store = store.NewStore()
	nodeMaster.votedLeader = ""

	nodeMaster.peers = make([]string, len(peers))
	copy(nodeMaster.peers, peers)
	nodeMaster.me = me

	return nodeMaster
}
