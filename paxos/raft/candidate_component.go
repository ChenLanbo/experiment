package raft

import (
	"time"
	"sync"
	"math/rand"

	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"flag"
	"log"
)

var (
	sleepTimeout = flag.Int("raft_candidate_sleep_timeout", 200, "")
)

type Candidate struct {
	mu sync.Mutex
	nodeMaster *NodeMaster
}

func (candidate *Candidate) Run() {
	candidate.VoteSelf()
	candidate.ProcessRequest()
}

func (candidate *Candidate) ProcessRequest() {
	go func() {
		for candidate.isCandidate() {
			candidate.ProcessOneRequest()
		}
	} ()
}

func (candidate *Candidate) ProcessOneRequest() {
	candidate.mu.Lock()
	defer candidate.mu.Unlock()

	op := candidate.nodeMaster.OpsQueue.Pull(time.Millisecond * time.Duration(*queuePullTimeout))

	if op.Request.AppendRequest != nil {
		reply := &pb.AppendReply{}

		if candidate.nodeMaster.store.CurrentTerm() <= *op.Request.AppendRequest.Term {
			// Append from the new leader
			candidate.nodeMaster.store.SetCurrentTerm(*op.Request.AppendRequest.Term)
			candidate.nodeMaster.votedLeader = "" // should be empty?
			candidate.nodeMaster.state = FOLLOWER
		}

		reply.Success = proto.Bool(false)
		reply.Term = proto.Uint64(candidate.nodeMaster.store.CurrentTerm())

		op.Callback <- *NewRaftReply(nil, reply)
	}

	if op.Request.VoteRequest != nil {
		reply := &pb.VoteReply{}

		if candidate.nodeMaster.store.CurrentTerm() >= *op.Request.VoteRequest.Term {
			// Reply false if term < currentTerm
			reply.Granted = proto.Bool(false)
			reply.Term = proto.Uint64(candidate.nodeMaster.store.CurrentTerm())
		} else  {
			// RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower
			candidate.nodeMaster.store.SetCurrentTerm(*op.Request.VoteRequest.Term)
			candidate.nodeMaster.state = FOLLOWER

			reply.Granted = proto.Bool(false)
			reply.Term = proto.Uint64(candidate.nodeMaster.store.CurrentTerm())
		}
		op.Callback <- *NewRaftReply(reply, nil)
	}
}

func (candidate *Candidate) VoteSelf() {
	go func() {
		for candidate.isCandidate() {
			candidate.VoteSelfOnce()
			if !candidate.isCandidate() {
				break
			}
			// Randomly sleep for some time
			time.Sleep(time.Millisecond * time.Duration(rand.Int31n(int32(*sleepTimeout)) + 1)) // TODO: make a flag
		}
	} ()
}

func (candidate *Candidate) VoteSelfOnce() {

	numSuccess := 0
	candidate.mu.Lock()
	defer candidate.mu.Unlock()

	log.Print("VoteSelfOnce\n")
	// 1. Increment current term
	candidate.nodeMaster.store.IncrementCurrentTerm()

	// 2. Send votes to peers
	// TODO: make this concurrent
	for _, peer := range(candidate.nodeMaster.peers) {
		if peer == candidate.nodeMaster.peers[candidate.nodeMaster.me] {
			continue
		}

		l := candidate.nodeMaster.store.Read(candidate.nodeMaster.store.LatestIndex())
		request := &pb.VoteRequest{
			Term:proto.Uint64(candidate.nodeMaster.store.CurrentTerm()),
			CandidateId:proto.String(candidate.nodeMaster.MyEndpoint()),
			LastLogIndex:proto.Uint64(*l.LogId),
			LastLogTerm:proto.Uint64(*l.Term)}

		reply, err := candidate.nodeMaster.Exchange.Vote(peer, request)
		if err != nil {
			log.Fatal(err, "\n")
			continue
		}

		if *reply.Granted {
			log.Println("Granted from peer", peer, "\n")
			numSuccess++
		} else {
			if candidate.nodeMaster.store.CurrentTerm() < *reply.Term {
				// RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower
				candidate.nodeMaster.store.SetCurrentTerm(*reply.Term)
				candidate.nodeMaster.state = FOLLOWER
				return
			}
		}

		if numSuccess + 1 > len(candidate.nodeMaster.peers) / 2 {
			// Guaranteed to be the new leader
			candidate.nodeMaster.state = LEADER
			break
		}
	}
}

func (candidate *Candidate) isCandidate() bool {
	candidate.mu.Lock()
	defer candidate.mu.Unlock()

	return candidate.nodeMaster.state == CANDIDATE
}

///////////////////////////////////////////////////////////////
// Constructor
///////////////////////////////////////////////////////////////

func NewCandidate(nodeMaster *NodeMaster) *Candidate {
	candidate := &Candidate{}
	candidate.nodeMaster = nodeMaster
	return candidate
}
