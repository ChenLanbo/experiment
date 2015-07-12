package paxos
import (
    "golang.org/x/net/context"
    pb "github.com/chenlanbo/experiment/paxos/protos"
)

type RaftServerState int
const (
    LEADER RaftServerState = iota + 1
    FOLLOWER
    CANDIDATE
)

type RaftServer struct {

}


func (raft *RaftServer) Vote(ctx context.Context, request *pb.VoteRequest) (*pb.VoteReply, error) {
    return nil, nil
}

func (raft *RaftServer) Append(ctx context.Context, request *pb.AppendRequest) (*pb.AppendReply, error) {
    return nil, nil
}