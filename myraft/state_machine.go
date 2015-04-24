package myraft

import "log"
import "sync/atomic"
import "time"

import myproto "github.com/chenlanbo/experiment/myraft/proto"
import "github.com/golang/protobuf/proto"

type Role int
const (
    Leader Role = iota + 1
    Follower
    Candidate
)

type StateMachine struct {
    master *NodeMaster

    // 
    stopped int32
    role Role
    vote_leader string
    current_term uint64
    last_appied uint64
    logs []myproto.Log
}

func NewStateMachine(s *NodeMaster) *StateMachine {
    sm := &StateMachine{}
    return sm
}

func (s *StateMachine) Start() {
    go func() {
        for !s.IsShutdown() {

            switch s.role {
            case Leader:
            case Follower:
            case Candidate:
            }
        }
        log.Println("StateMachine stoppped")
    } ()
}

func (s *StateMachine) RunLeader() {
    go func() {
        // Start to accept client's request
    } ()

    for !s.IsShutdown() {
        select {
        case cb := <-s.master.peer_server.VoteChan():

            rep := s.HandleVote(cb.Req)
            cb.Rep <- rep
            if rep.GetGranted() {
                // I am not the leader anymore, switch to follower
                s.current_term = cb.Req.GetTerm()
                s.vote_leader = cb.Req.GetCandidateId()
                s.role = Follower
                return
            }

        case cb := <-s.master.peer_server.AppendChan():
            //
        case <-time.After(ChannelWaitTimeout):
            //
        }
    }
}

func (s *StateMachine) RunFollower() {
    for !s.IsShutdown() {
        select {
        case cb := <-s.master.peer_server.VoteChan():
            // Handle vote request
            rep := s.HandleVote(cb.Req)
            cb.Rep <- rep
            if rep.GetGranted() {
                // I am still a follower
                s.current_term = cb.Req.GetTerm()
                s.vote_leader = cb.Req.GetCandidateId()
                s.role = Follower
                return
            }

        case cb := <-s.master.peer_server.AppendChan():
            //
        case <-time.After(ChannelWaitTimeout):
            // May not receive append from the leader, switch to candidate
            s.role = Candidate
            break
        }
    }
}

func (s *StateMachine) Candidate() {

    go func() {
        // Have another routine accept incoming messages
        for !s.IsShutdown() {
            select {
            // Here do not accept vote requests from other peers (from paper).
            case cb := <-s.master.peer_server.AppendChan():
                // Another peer may have become the leader, get me out of
                // candidate state.
            case <-time.After(time.Second):
                // Do nothing
            }
        }
    } ()

    for {
        // Start leader election, vote my self
        new_term := s.IncrementCurrentTerm()
        s.vote_leader = s.master.Addr()

        last_log := s.logs[s.last_appied]
        req := myproto.VoteRequest {
            Term: proto.Uint64(new_term),
            CandidateId: proto.String(s.vote_leader),
            LastLogIndex: proto.Uint64(last_log.GetLogId()),
            LastLogTerm: proto.Uint64(last_log.GetTerm()),
        }

    }
}

func (s *StateMachine) HandleVote(req *myproto.VoteRequest) myproto.VoteReply {
    rep := myproto.VoteReply{}

    if req.GetTerm() < s.current_term {
        rep.CurrentTerm = proto.Uint64(s.current_term)
        rep.Granted = proto.Bool(false)
    } else if s.vote_leader == "" {
        // I didn't vote any leader
        rep.CurrentTerm = proto.Uint64(req.GetTerm())
        rep.Granted = proto.Bool(true)
    } else {
        last_log := s.logs[s.last_appied]
        if req.GetLastLogIndex() >= last_log.GetLogId() &&
           req.GetLastLogTerm() >= last_log.GetTerm() {
           // candidate's log is at least as up-to-date as receiver's log, granted
           rep.CurrentTerm = proto.Uint64(req.GetTerm())
           rep.Granted = proto.Bool(true)
        } else {
            rep.CurrentTerm = proto.Uint64(s.current_term)
            rep.Granted = proto.Bool(false)
        }
    }

    return rep
}

func (s *StateMachine) HandleAppend(req *myproto.AppendRequest) myproto.AppendReply {
    rep := myproto.AppendReply{}

    return rep
}

func (s *StateMachine) Shutdown() {
    atomic.StoreInt32(&s.stopped, 1)
}

func (s *StateMachine) IsShutdown() bool {
    return atomic.LoadInt32(&s.stopped) != 0
}

func (s *StateMachine) GetCurrentTerm() uint64 {
    return atomic.LoadUint64(&s.current_term)
}

func (s *StateMachine) IncrementCurrentTerm() uint64 {
    return atomic.AddUint64(&s.current_term, 1)
}

func (s *StateMachine) UpdateCurrentTerm(new_term uint64) {
    atomic.StoreUint64(&s.current_term, new_term)
}
