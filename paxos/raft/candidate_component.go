package raft

import (
	"time"
	"sync"
	"math/rand"

	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"flag"
	"log"
	"sync/atomic"
)

var (
	sleepTimeout = flag.Int("raft_candidate_sleep_timeout", 200, "")
)

type Candidate struct {
	mu sync.Mutex
	nodeMaster *NodeMaster

	newTermChan chan uint64
	votedChan chan bool

	stopped int32
}

func (candidate *Candidate) Run() {

	for !candidate.Stopped() {
		candidate.nodeMaster.store.IncrementCurrentTerm()
		candidate.nodeMaster.votedLeader = ""
		newTerm := candidate.nodeMaster.store.CurrentTerm()
		l := candidate.nodeMaster.store.Read(candidate.nodeMaster.store.LatestIndex())

		processor := NewCandidateRequestProcessor(candidate)
		voter := NewCandidateVoter(candidate)

		voter.VoteSelfAtTerm(newTerm, l.GetTerm(), l.GetLogId())
		processor.ProcessRequestsAtTerm(newTerm)

		select {
		case higherTerm := <- candidate.newTermChan:
			if higherTerm >= newTerm {
				candidate.nodeMaster.state = FOLLOWER
				processor.Stop()
				voter.Stop()
				return
			} else {
				// shouldn't happend
			}
		case result := <- candidate.votedChan:
			if result {
				// I am the new leader
				candidate.nodeMaster.state = LEADER
				processor.Stop()
				voter.Stop()
				return
			} else {
				log.Println(candidate.nodeMaster.MyEndpoint(), "no leader elected at term", newTerm)
				time.Sleep(time.Millisecond * time.Duration(rand.Int31n(int32(*sleepTimeout)) + 1))
				processor.Stop()
				voter.Stop()
			}
		}
	}
}

func (candidate *Candidate) Stop() {
	atomic.StoreInt32(&candidate.stopped, 1)
	candidate.votedChan <- true
}

func (candidate *Candidate) Stopped() bool {
	return atomic.LoadInt32(&candidate.stopped) == 1
}

///////////////////////////////////////////////////////////////
// CandidateRequestProcessor
///////////////////////////////////////////////////////////////

type CandidateRequestProcessor struct {
	candidate *Candidate
	stopped int32
	mu sync.Mutex
}

func (processor *CandidateRequestProcessor) Stop() {
	atomic.StoreInt32(&processor.stopped, 1)
}

func (processor *CandidateRequestProcessor) Stopped() bool {
	return atomic.LoadInt32(&processor.stopped) == 1
}

func (processor *CandidateRequestProcessor) ProcessRequestsAtTerm(newTerm uint64) {
	go func() {
		processor.mu.Lock()
		defer processor.mu.Unlock()

		for {
			if processor.Stopped() {
				return
			}

			op := processor.candidate.nodeMaster.OpsQueue.Pull(
				time.Millisecond * time.Duration(*queuePullTimeout))

			if op == nil {
				continue
			}

			if op.Request.AppendRequest != nil {
				reply := &pb.AppendReply{
					Success:proto.Bool(false),
					Term:proto.Uint64(newTerm)}

				op.Callback <- *NewRaftReply(nil, reply, nil)

				if newTerm <= op.Request.AppendRequest.GetTerm() {
					processor.candidate.newTermChan <- op.Request.AppendRequest.GetTerm()
					return
				}

			} else if op.Request.VoteRequest != nil {

				reply := &pb.VoteReply{
					Granted:proto.Bool(false),
					Term:proto.Uint64(newTerm)}
				op.Callback <- *NewRaftReply(reply, nil, nil)

				if newTerm < op.Request.VoteRequest.GetTerm() {
					processor.candidate.newTermChan <- op.Request.VoteRequest.GetTerm()
					return
				}

			} else if op.Request.PutRequest != nil {
				reply := &pb.PutReply{Success:proto.Bool(false)}
				op.Callback <- *NewRaftReply(nil, nil, reply)
			}
		}
	} ()
}

///////////////////////////////////////////////////////////////
// CandidateVoter
///////////////////////////////////////////////////////////////

type CandidateVoter struct {
	candidate *Candidate
	stopped int32
	mu sync.Mutex
}

func (voter *CandidateVoter) Stop() {
	atomic.StoreInt32(&voter.stopped, 1)
}

func (voter *CandidateVoter) Stopped() bool {
	return atomic.LoadInt32(&voter.stopped) == 1
}

func (voter *CandidateVoter) VoteSelfAtTerm(newTerm, lastLogTerm, lastLogIndex uint64) {
	go func() {
		voter.mu.Lock()
		defer voter.mu.Unlock()
		if voter.Stopped() {
			return
		}

		peers := voter.candidate.nodeMaster.PeerEndpoints()
		shuffle := rand.Perm(len(peers))
		shuffledPeers := make([]string, len(peers))
		for i, _ := range shuffledPeers {
			shuffledPeers[i] = peers[shuffle[i]]
		}

		request := &pb.VoteRequest{
			Term:proto.Uint64(newTerm),
			CandidateId:proto.String(voter.candidate.nodeMaster.MyEndpoint()),
			LastLogTerm:proto.Uint64(lastLogTerm),
			LastLogIndex:proto.Uint64(lastLogIndex)}

		numSuccess := 1

		for _, peer := range shuffledPeers {
			if voter.Stopped() {
				return
			}

			reply, err := voter.candidate.nodeMaster.Exchange.Vote(peer, request)
			if err != nil {
				log.Println(err)
				continue
			}

			if reply.GetGranted() {
				log.Println("Granted from", peer)
				numSuccess++
			} else {
				log.Println(
					voter.candidate.nodeMaster.MyEndpoint(),
					"not granted from", peer,
					"my term", newTerm,
					"reply term", reply.GetTerm())
				if newTerm < reply.GetTerm() {
					// RPC request or response contains term T > currentTerm.
					// Signal Candidate to become follower to set new term
					voter.candidate.newTermChan <- reply.GetTerm()
					return
				}
			}

			if numSuccess > (len(peers) + 1) / 2 {
				// Guaranteed to be the new leader
				voter.candidate.votedChan <- true
				return
			}
		}

		if numSuccess > (len(peers) + 1) / 2 {
			// Guaranteed to be the new leader
			voter.candidate.votedChan <- true
			return
		}

		// Leader not guaranteed
		voter.candidate.votedChan <- false
	} ()
}

///////////////////////////////////////////////////////////////
// Constructor
///////////////////////////////////////////////////////////////

func NewCandidateRequestProcessor(candidate *Candidate) *CandidateRequestProcessor {
	processor := &CandidateRequestProcessor{}
	processor.candidate = candidate
	processor.stopped = 0
	return processor
}

func NewCandidateVoter(candidate *Candidate) *CandidateVoter {
	voter := &CandidateVoter{}
	voter.candidate = candidate
	voter.stopped = 0
	return voter
}

func NewCandidate(nodeMaster *NodeMaster) *Candidate {
	candidate := &Candidate{}
	candidate.nodeMaster = nodeMaster
	candidate.votedChan = make(chan bool, 3)
	candidate.newTermChan = make(chan uint64, 3)
	return candidate
}