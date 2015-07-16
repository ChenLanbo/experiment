package raft
import "sync"

type Leader struct {
	mu sync.Mutex
	nodeMaster *NodeMaster
}

func (leader *Leader) Run() {

}

func (leader *Leader) ProcessRequest() {

}

func (leader *Leader) isLeader() bool {
	leader.mu.Lock()
	defer leader.mu.Unlock()

	return leader.nodeMaster.state == LEADER
}