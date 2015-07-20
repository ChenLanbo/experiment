package raft

import (
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"google.golang.org/grpc"
	"log"
	"golang.org/x/net/context"
)

func SendVote(peer string, request *pb.VoteRequest) (*pb.VoteReply, error) {
	conn, err := grpc.Dial(peer)
	if err != nil {
		log.Fatal("")
		return nil, err
	}

	defer conn.Close()
	c := pb.NewRaftServerClient(conn)

	reply, err1 := c.Vote(context.Background(), request)
	if err1 != nil {
		log.Fatal(err1, "\n")
		return nil, err1
	}

	return reply, nil
}

func SendAppend(peer string, request *pb.AppendRequest) (*pb.AppendReply, error) {
	conn, err := grpc.Dial(peer)
	if err != nil {
		log.Fatal("")
		return nil, err
	}

	defer conn.Close()
	c := pb.NewRaftServerClient(conn)

	reply, err1 := c.Append(context.Background(), request)
	if err1 != nil {
		log.Fatal("")
		return nil, err1
	}

	return reply, nil
}
