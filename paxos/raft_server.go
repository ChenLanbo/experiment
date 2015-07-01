package paxos

import "log"
import "net"
import "net/rpc"
import "sync"

import raftproto "github.com/chenlanbo/experiment/paxos/proto"
import "github.com/golang/protobuf/proto"

type State int
const (
    Leader State = iota + 1
    Follower
    Candidate
)

type VoteCallback struct {
    req *raftproto.VoteRequest
    callback chan *raftproto.VoteReply
}

type AppendCallback struct {
    req *raftproto.AppendRequest
    callback chan *raftproto.AppendReply
}

type RaftServer struct {
    state State
    vote_leader string
    current_term uint64
    commit_index uint64
    last_applied uint64
    logs []raftproto.Log

    peers []string
    me int

    vote_chan chan VoteCallback
    append_chan chan AppendCallback

    mu sync.Mutex
}

func (raft *RaftServer) Vote(req *raftproto.VoteRequest, rep *raftproto.VoteReply) error {
    rep.Granted = proto.Bool(true)
    rep.Term = proto.Uint64(0)
    return nil
}

func (raft *RaftServer) Append(req *raftproto.AppendRequest, rep *raftproto.AppendReply) error {
    rep.Success = proto.Bool(true)
    rep.Term = proto.Uint64(0)
    return nil
}

func (raft *RaftServer) Start() {
    // Open socket
    rpc.Register(raft);
    l, e := net.Listen("tcp", raft.peers[raft.me])
    if e != nil {
        log.Fatal("listen error: ", e)
    }

    // Start the state machine
    raft.StartStateMachine()

    // Start to accept connections
    go func() {
        for !raft.Stopped() {
            if conn, err := l.Accept(); err != nil {
                log.Fatal("accept error: ", e)
            } else {
                go rpc.ServeConn(conn)
            }
        }
    } ()
}

func (raft *RaftServer) Stop() {
}

func (raft *RaftServer) Stopped() bool {
    return false
}

func (raft *RaftServer) StartStateMachine() {
    go func() {
        for !raft.Stopped() {
            switch raft.state {
            case Leader:
                raft.RunLeader()
            case Follower:
                raft.RunFollower()
            case Candidate:
                raft.RunCandidate()
            }
        }
    } ()
}

func (raft *RaftServer) RunLeader() {
}

func (raft *RaftServer) RunFollower() {
    for !raft.Stopped() {
        select {
        }
    }
}

func (raft *RaftServer) RunCandidate() {
}

func NewRaftServer(peers []string, me int) *RaftServer {
    raft := &RaftServer{}

    raft.peers = make([]string, len(peers))
    copy(raft.peers, peers)
    raft.me = me

    raft.state = Follower
    raft.vote_leader = ""
    raft.current_term = 0
    raft.commit_index = 0
    raft.last_applied = 0
    raft.logs = make([]raftproto.Log, 0)

    raft.vote_chan = make(chan VoteCallback, 2)
    raft.append_chan = make(chan AppendCallback, 2)

    return raft
}

