package raft
import (
    "golang.org/x/net/context"
    pb "github.com/chenlanbo/experiment/paxos/protos"
    "github.com/golang/protobuf/proto"
)


type RaftServer struct {}

func (raft *RaftServer) Vote(ctx context.Context, request *pb.VoteRequest) (*pb.VoteReply, error) {
    reply := &pb.VoteReply{}
    reply.Term = proto.Uint64(1)
    reply.Granted = proto.Bool(false)
    return reply, nil
}

func (raft *RaftServer) Append(ctx context.Context, request *pb.AppendRequest) (*pb.AppendReply, error) {
    return nil, nil
}

func (raft *RaftServer) Put(ctx context.Context, request *pb.PutRequest) (*pb.PutReply, error) {
    reply := &pb.PutReply{}
    reply.Success = proto.Bool(true)
    reply.LeaderId = proto.String("abc")
    return reply, nil
}