package raft
import (
    "golang.org/x/net/context"
    pb "github.com/chenlanbo/experiment/paxos/protos"
    "errors"
    "github.com/chenlanbo/experiment/paxos/raft/rpc"
    "net"
    "log"
    "google.golang.org/grpc"
    "sync/atomic"
    "sync"
)

// mockgen github.com/chenlanbo/experiment/paxos/raft RaftServer > mock_raft/mock_raft_server.go
type RaftServer struct {
    nodeMaster *NodeMaster
    stopped int32
    l net.Listener
    mu sync.Mutex
}

func (raft *RaftServer) Start() {
    go func () {
        raft.mu.Lock()
        defer raft.mu.Unlock()

        if raft.Stopped() {
            return
        }

        log.Println(raft.nodeMaster.MyEndpoint(), "start serving requests")
        s := grpc.NewServer()
        pb.RegisterRaftServer(s, raft)
        s.Serve(raft.l)
        s.Stop()
    } ()

    raft.nodeMaster.Start()
}

func (raft *RaftServer) Stop() {
    atomic.StoreInt32(&raft.stopped, 1)
    raft.l.Close()
    raft.nodeMaster.Stop()
}

func (raft *RaftServer) Stopped() bool {
    return atomic.LoadInt32(&raft.stopped) == 1
}

func (raft *RaftServer) Vote(ctx context.Context, request *pb.VoteRequest) (*pb.VoteReply, error) {
    op := NewRaftOperation(NewRaftRequest(request, nil, nil))
    defer close(op.Callback)
    if raft.Stopped() {
        return nil, errors.New("Server stopped")
    }

    raft.nodeMaster.OpsQueue.Push(op)

    reply := <- op.Callback
    if reply.VoteReply == nil {
        return nil, errors.New("Internal Server Error")
    }
    return reply.VoteReply, nil
}

func (raft *RaftServer) Append(ctx context.Context, request *pb.AppendRequest) (*pb.AppendReply, error) {
    op := NewRaftOperation(NewRaftRequest(nil, request, nil))
    defer close(op.Callback)
    if raft.Stopped() {
        return nil, errors.New("Server stopped")
    }

    raft.nodeMaster.OpsQueue.Push(op)

    reply := <- op.Callback
    if reply.AppendReply == nil {
        return nil, errors.New("Internal Server Error")
    }
    return reply.AppendReply, nil
}

func (raft *RaftServer) Put(ctx context.Context, request *pb.PutRequest) (*pb.PutReply, error) {
    op := NewRaftOperation(NewRaftRequest(nil, nil, request))
    defer close(op.Callback)
    if raft.Stopped() {
        return nil, errors.New("Server stopped")
    }

    raft.nodeMaster.OpsQueue.Push(op)

    reply := <- op.Callback
    if reply.PutReply == nil {
        return nil, errors.New("Internal Server Error")
    }
    return reply.PutReply, nil
}

///////////////////////////////////////////////////////////////
// Constructor
///////////////////////////////////////////////////////////////

func NewRaftServer(peers []string, me int) *RaftServer {
    raft := &RaftServer{}

    master := NewNodeMaster(rpc.NewMessageExchange(), peers, me)
    raft.nodeMaster = master
    raft.stopped = 0

    l, err := net.Listen("tcp", peers[me])
    if err != nil {
        log.Println(err)
        return nil
    }
    raft.l = l

    return raft
}
