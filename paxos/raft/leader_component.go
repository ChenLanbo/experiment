package raft
import (
	"sync"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"log"
	"time"
	"sort"
	"sync/atomic"
)

///////////////////////////////////////////////////////////////
// Leader
///////////////////////////////////////////////////////////////

type Leader struct {
	mu sync.Mutex
	nodeMaster *NodeMaster
	replicators []*LogReplicator
	committer *LogCommitter
	processor *LeaderRequestProcessor
	newTermChan chan uint64
}

func (leader *Leader) Run() {
	leader.processor.Start()
	leader.committer.Start()
	for _, replicator := range leader.replicators {
		replicator.Start()
	}

	for {
		newTerm := <- leader.newTermChan
		if newTerm > leader.nodeMaster.store.CurrentTerm() {
			leader.nodeMaster.state = FOLLOWER

			leader.processor.Stop()
			leader.committer.Stop()
			for _, replicator := range leader.replicators {
				replicator.Stop()
			}

			return
		}
	}
}

///////////////////////////////////////////////////////////////
// LeaderRequestProcessor
///////////////////////////////////////////////////////////////

type LeaderRequestProcessor struct {
	leader *Leader
	stopped int32
	mu sync.Mutex
}

func (processor *LeaderRequestProcessor) Start() {
	go func() {
		processor.mu.Lock()
		defer processor.mu.Unlock()

		for !processor.Stopped() {
			if !processor.ProcessOnce() {
				return
			}
		}
	} ()
}

func (processor *LeaderRequestProcessor) ProcessOnce() bool {
	queue := processor.leader.nodeMaster.OpsQueue
	store := processor.leader.nodeMaster.store

	op := queue.Pull(time.Millisecond * time.Duration(*queuePullTimeout))
	if op == nil {
		return true
	}

	if op.Request.AppendRequest != nil {
		// Process append request

		reply := &pb.AppendReply{}
		reply.Success = proto.Bool(false)
		reply.Term = proto.Uint64(store.CurrentTerm())

		op.Callback <- *NewRaftReply(nil, reply, nil)

		if store.CurrentTerm() < op.Request.AppendRequest.GetTerm() {
			// I am not leader anymore
			processor.leader.newTermChan <- op.Request.AppendRequest.GetTerm()
			return false
		}

	} else if op.Request.VoteRequest != nil {
		// Process vote request

		reply := &pb.VoteReply{}
		reply.Granted = proto.Bool(false)
		reply.Term = proto.Uint64(store.CurrentTerm())

		op.Callback <- *NewRaftReply(reply, nil, nil)

		if store.CurrentTerm() < op.Request.AppendRequest.GetTerm() {
			// I am not leader anymore
			processor.leader.newTermChan <- op.Request.AppendRequest.GetTerm()
			return false
		}

	} else if op.Request.PutRequest != nil {
		// Process put request

		// Write data to store and put into inflight requests.
		// Once commit index advanced by committer, ack the request.
		logId := store.WriteKeyValue(
			*op.Request.PutRequest.Key, op.Request.PutRequest.Value)
		processor.leader.nodeMaster.inflightRequests[logId] = op
	}

	return true
}

func (processor *LeaderRequestProcessor) Stop() {
	atomic.StoreInt32(&processor.stopped, 1)
}

func (processor *LeaderRequestProcessor) Stopped() bool {
	return atomic.LoadInt32(&processor.stopped) == 1
}

///////////////////////////////////////////////////////////////
// Log replicator
///////////////////////////////////////////////////////////////

type LogReplicator struct {
	leader *Leader
	peer string
	replicateIndex uint64
	prevLog *pb.Log
	stopped int32
	mu sync.Mutex
}

func (replicator *LogReplicator) Start() {
	go func() {
		replicator.mu.Lock()
		defer replicator.mu.Unlock()

		for !replicator.Stopped() {
			if !replicator.ReplicateOnce() {
				return
			}
		}
	} ()
}

func (replicator *LogReplicator) ReplicateOnce() bool {
	store := replicator.leader.nodeMaster.store

	newLog := store.Poll(replicator.replicateIndex + 1)
	logsToReplicate := make([]*pb.Log, 0)
	if newLog != nil {
		logsToReplicate = append(logsToReplicate, newLog)
	}

	request := &pb.AppendRequest{
		Term:proto.Uint64(store.CurrentTerm()),
		LeaderId:proto.String(replicator.leader.nodeMaster.MyEndpoint()),
		PrevLogTerm:proto.Uint64(replicator.prevLog.GetTerm()),
		PrevLogIndex:proto.Uint64(replicator.prevLog.GetLogId()),
		CommitIndex:proto.Uint64(store.CommitIndex()),
		Logs:logsToReplicate}

	reply, err := replicator.leader.nodeMaster.Exchange.Append(replicator.peer, request)
	if err != nil {
		log.Fatal("Failed to replicate logs to", replicator.peer)
	} else {
		if reply.GetSuccess() {
			if len(logsToReplicate) > 0 {
				replicator.replicateIndex++
				replicator.prevLog = newLog
				// log.Println("Replicator", replicator.peer, ":", replicator.replicateIndex)
			}
		} else {
			if reply.GetTerm() > store.CurrentTerm() {
				// Return to follower state
				replicator.leader.newTermChan <- reply.GetTerm()
				return false
			} else {
				// Replicate backward
				if replicator.replicateIndex > 0 {
					replicator.replicateIndex--
				}
				replicator.prevLog = store.Read(replicator.replicateIndex)
			}
		}
	}

	return true
}

func (replicator *LogReplicator) Stop() {
	atomic.StoreInt32(&replicator.stopped, 1)
}

func (replicator *LogReplicator) Stopped() bool {
	return atomic.LoadInt32(&replicator.stopped) == 1
}

///////////////////////////////////////////////////////////////
// Replicate index sorter
///////////////////////////////////////////////////////////////

type ReplicateIndexSorter []uint64

func (s ReplicateIndexSorter) Len() int {
	return len(s)
}

func (s ReplicateIndexSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ReplicateIndexSorter) Less(i, j int) bool {
	return s[i] > s[j]
}

///////////////////////////////////////////////////////////////
// Log committer
///////////////////////////////////////////////////////////////

type LogCommitter struct {
	leader *Leader
	stopped int32
	mu sync.Mutex
}

func (committer *LogCommitter) Start() {
	go func() {
		committer.mu.Lock()
		defer committer.mu.Unlock()

		for !committer.Stopped() {
			committer.CommitOnce()
		}
	} ()
}

func (committer *LogCommitter) CommitOnce() {
	store := committer.leader.nodeMaster.store
	updated := false
	if store.CommitIndex() < store.LatestIndex() {
		arr := make([]uint64, len(committer.leader.replicators) + 1)
		for id, replicator := range(committer.leader.replicators) {
			arr[id] = replicator.replicateIndex
		}
		arr[len(arr) - 1] = store.LatestIndex()

		sort.Sort(ReplicateIndexSorter(arr))
		if arr[len(arr) / 2] > store.CommitIndex() {
			store.SetCommitIndex(arr[len(arr) / 2])
			updated = true
		} else {
			time.Sleep(time.Millisecond * 10)
		}
	} else {
		time.Sleep(time.Millisecond * 10)
	}

	// Ack inflight put requests
	if updated {
		reply := &pb.PutReply{Success:proto.Bool(true)}
		commitIndex := committer.leader.nodeMaster.store.CommitIndex()
		for logId, op := range committer.leader.nodeMaster.inflightRequests {
			if logId <= commitIndex {
				op.Callback <- *NewRaftReply(nil, nil, reply)
			}
			delete(committer.leader.nodeMaster.inflightRequests, logId)
		}
	}
}

func (committer *LogCommitter) Stop() {
	atomic.StoreInt32(&committer.stopped, 1)
}

func (committer *LogCommitter) Stopped() bool {
	return atomic.LoadInt32(&committer.stopped) == 1
}

///////////////////////////////////////////////////////////////
// Constructors
///////////////////////////////////////////////////////////////

func NewLeaderRequestProcessor(leader *Leader) *LeaderRequestProcessor {
	processor := &LeaderRequestProcessor{}
	processor.leader = leader
	processor.stopped = 0
	return processor
}

func NewLogReplicator(leader *Leader, peer string) *LogReplicator {
	replicator := &LogReplicator{}
	replicator.leader = leader
	replicator.peer = peer

	// Start with commit index
	replicator.replicateIndex = leader.nodeMaster.store.CommitIndex()
	replicator.prevLog = leader.nodeMaster.store.Read(replicator.replicateIndex)

	replicator.stopped = 0

	return replicator
}

func NewLogCommitter(leader *Leader) *LogCommitter {
	committer := &LogCommitter{}
	committer.leader = leader
	committer.stopped = 0
	return committer
}

func NewLeader(nodeMaster *NodeMaster) *Leader {
	leader := &Leader{}
	leader.nodeMaster = nodeMaster

	logReplicators := make([]*LogReplicator, 0)
	for _, peer := range(nodeMaster.PeerEndpoints()) {
		logReplicators = append(logReplicators, NewLogReplicator(leader, peer))
	}
	leader.replicators = logReplicators

	leader.committer = NewLogCommitter(leader)
	leader.processor = NewLeaderRequestProcessor(leader)

	leader.newTermChan = make(chan uint64, len(logReplicators) + 1)

	return leader
}
