package raft
import (
    "golang.org/x/net/context"
    pb "github.com/chenlanbo/experiment/paxos/protos"
    "github.com/golang/protobuf/proto"
)

// mockgen github.com/chenlanbo/experiment/paxos/raft RaftServer > mock_raft/mock_raft_server.go
type RaftServer interface {

    Vote(ctx context.Context, request *pb.VoteRequest) (*pb.VoteReply, error)

    Append(ctx context.Context, request *pb.AppendRequest) (*pb.AppendReply, error)

    Put(ctx context.Context, request *pb.PutRequest) (*pb.PutReply, error)
}

type RaftServerImpl struct {}

func (raft *RaftServerImpl) Vote(ctx context.Context, request *pb.VoteRequest) (*pb.VoteReply, error) {
    reply := &pb.VoteReply{}
    reply.Term = proto.Uint64(1)
    reply.Granted = proto.Bool(false)
    return reply, nil
}

func (raft *RaftServerImpl) Append(ctx context.Context, request *pb.AppendRequest) (*pb.AppendReply, error) {
    return nil, nil
}

func (raft *RaftServerImpl) Put(ctx context.Context, request *pb.PutRequest) (*pb.PutReply, error) {
    reply := &pb.PutReply{}
    reply.Success = proto.Bool(true)
    reply.LeaderId = proto.String("abc")
    return reply, nil
}