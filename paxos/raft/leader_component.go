package raft
import (
	"sync"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"log"
	"time"
	"sort"
)

///////////////////////////////////////////////////////////////
// Leader
///////////////////////////////////////////////////////////////

type Leader struct {
	mu sync.Mutex
	nodeMaster *NodeMaster
	logReplicators []*LogReplicator
	logCommitter *LogCommitter
}

func (leader *Leader) Run() {
	leader.ProcessRequest()
	leader.ReplicateLog()
	leader.CommitLog()
}

func (leader *Leader) ProcessRequest() {
	go func() {
		for leader.isLeader() {
			leader.ProcessOneRequest()
		}
	} ()
}

func (leader *Leader) ProcessOneRequest() {

	op := leader.nodeMaster.OpsQueue.Pull(time.Millisecond * time.Duration(*queuePullTimeout))
	if op == nil {
		return
	}

	if op.Request.AppendRequest != nil {
		// Process append request

		reply := &pb.AppendReply{}
		reply.Success = proto.Bool(false)
		reply.Term = proto.Uint64(leader.nodeMaster.store.CurrentTerm())

		if leader.nodeMaster.store.CommitIndex() < *op.Request.AppendRequest.Term {
			// I am not leader anymore
			leader.nodeMaster.store.SetCurrentTerm(*op.Request.AppendRequest.Term)
			leader.Revoke()
		}

		op.Callback <- *NewRaftReply(nil, reply, nil)

	} else if op.Request.VoteRequest != nil {
		// Process vote request

		reply := &pb.VoteReply{}
		reply.Granted = proto.Bool(false)
		reply.Term = proto.Uint64(leader.nodeMaster.store.CurrentTerm())

		if leader.nodeMaster.store.CommitIndex() < *op.Request.AppendRequest.Term {
			// I am not leader anymore
			leader.nodeMaster.store.SetCurrentTerm(*op.Request.AppendRequest.Term)
			leader.Revoke()
		}

		op.Callback <- *NewRaftReply(reply, nil, nil)

	} else if op.Request.PutRequest != nil {
		// Process put request

		// Write data to store and put into inflight requests.
		// Once commit index advanced by committer, ack the request.
		logId := leader.nodeMaster.store.WriteKeyValue(
			*op.Request.PutRequest.Key, op.Request.PutRequest.Value)
		leader.nodeMaster.inflightRequests[logId] = op
	}
}

func (leader *Leader) ReplicateLog() {
	for _, replicator := range(leader.logReplicators) {
		replicator.Start()
	}
}

func (leader *Leader) CommitLog() {
	leader.logCommitter.Start()
}

// Revoke the leader itself as a leader.
func (leader *Leader) Revoke() {
	leader.mu.Lock()
	defer leader.mu.Unlock()

	if leader.nodeMaster.state != LEADER {
		return
	}

	for _, r := range(leader.logReplicators) {
		r.Stop()
	}
	<- time.After(time.Second)
	leader.nodeMaster.state = FOLLOWER
}

func (leader *Leader) isLeader() bool {
	leader.mu.Lock()
	defer leader.mu.Unlock()

	return leader.nodeMaster.state == LEADER
}

///////////////////////////////////////////////////////////////
// Log replicator
///////////////////////////////////////////////////////////////

type LogReplicator struct {
	parent *Leader
	peer string
	replicateIndex uint64
	prevLog *pb.Log
	stopped bool
}

func (replicator *LogReplicator) Start() {
	go func() {
		for !replicator.stopped {
			replicator.ReplicateOnce()
		}
	} ()
}

func (replicator *LogReplicator) ReplicateOnce() {
	store := replicator.parent.nodeMaster.store

	newLog := store.Poll(replicator.replicateIndex + 1)
	logsToReplicate := make([]*pb.Log, 0)
	if newLog != nil {
		logsToReplicate = append(logsToReplicate, newLog)
	}

	request := &pb.AppendRequest{
		Term:proto.Uint64(store.CurrentTerm()),
		LeaderId:proto.String(replicator.parent.nodeMaster.MyEndpoint()),
		PrevLogTerm:proto.Uint64(*replicator.prevLog.Term),
		PrevLogIndex:proto.Uint64(*replicator.prevLog.LogId),
		CommitIndex:proto.Uint64(store.CommitIndex()),
		Logs:logsToReplicate}

	reply, err := replicator.parent.nodeMaster.Exchange.Append(replicator.peer, request)
	if err != nil {
		log.Fatal("Failed to replicate logs to", replicator.peer)
	} else {
		if *reply.Success {
			if len(logsToReplicate) > 0 {
				replicator.replicateIndex++
				replicator.prevLog = newLog
			}
		} else {
			if *reply.Term > store.CurrentTerm() {
				// return to follower state
			} else {
				if replicator.replicateIndex > 0 {
					replicator.replicateIndex--
				}
				replicator.prevLog = store.Read(replicator.replicateIndex)
			}
		}
	}
}

func (replicator *LogReplicator) Stop() {
	replicator.stopped = true
}

///////////////////////////////////////////////////////////////
// Log committer
///////////////////////////////////////////////////////////////

type LogCommitter struct {
	leader *Leader
	stopped bool
}

type ReplicateIndexSorter []uint64

func (s ReplicateIndexSorter) Len() int {
	return len(s)
}

func (s ReplicateIndexSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ReplicateIndexSorter) Less(i, j int) bool {
	return s[i] < s[j]
}

func (committer *LogCommitter) Start() {
	go func() {
		for !committer.stopped {
			committer.CommitOnce()
		}
	} ()
}

func (committer *LogCommitter) CommitOnce() {
	if committer.leader.nodeMaster.store.CommitIndex() < committer.leader.nodeMaster.store.LatestIndex() {
		arr := make([]uint64, len(committer.leader.logReplicators))
		for id, replicator := range(committer.leader.logReplicators) {
			arr[id] = replicator.replicateIndex
		}

		sort.Sort(ReplicateIndexSorter(arr))
		log.Println("ARR:", arr, "setting commit index to", arr[len(arr) / 2])
		committer.leader.nodeMaster.store.SetCommitIndex(arr[len(arr) / 2])
	} else {
		<- time.After(time.Millisecond * 10)
	}

	// Ack inflight put requests
	reply := &pb.PutReply{Success:proto.Bool(true)}
	commitIndex := committer.leader.nodeMaster.store.CommitIndex()
	for logId, op := range committer.leader.nodeMaster.inflightRequests {
		if logId <= commitIndex {
			op.Callback <- *NewRaftReply(nil, nil, reply)
		}
		delete(committer.leader.nodeMaster.inflightRequests, logId)
	}
}

///////////////////////////////////////////////////////////////
// Constructors
///////////////////////////////////////////////////////////////

func NewLogReplicator(leader *Leader, peer string) *LogReplicator {
	replicator := &LogReplicator{}
	replicator.parent = leader
	replicator.peer = peer

	// Start with commit index
	replicator.replicateIndex = leader.nodeMaster.store.CommitIndex()
	replicator.prevLog = leader.nodeMaster.store.Read(replicator.replicateIndex)

	replicator.stopped = false

	return replicator
}

func NewLogCommitter(leader *Leader) *LogCommitter {
	committer := &LogCommitter{}
	committer.leader = leader
	committer.stopped = false
	return committer
}

func NewLeader(nodeMaster *NodeMaster) *Leader {
	leader := &Leader{}
	leader.nodeMaster = nodeMaster

	logReplicators := make([]*LogReplicator, 0)
	for _, peer := range(nodeMaster.PeerEndpoints()) {
		logReplicators = append(logReplicators, NewLogReplicator(leader, peer))
	}
	leader.logReplicators = logReplicators

	leader.logCommitter = NewLogCommitter(leader)

	return leader
}
