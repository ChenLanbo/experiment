package rpc

import (
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"google.golang.org/grpc"
	"log"
	"golang.org/x/net/context"
)

type MessageExchange interface {

	Vote(peer string, request *pb.VoteRequest) (*pb.VoteReply, error)

	Append(peer string, request *pb.AppendRequest) (*pb.AppendReply, error)

	Put(peer string, request *pb.PutRequest) (*pb.PutReply, error)
}

type MessageExchangeImpl struct {}

func (hub MessageExchangeImpl) Vote(peer string, request *pb.VoteRequest) (*pb.VoteReply, error) {
	conn, err := grpc.Dial(peer)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer conn.Close()

	c := pb.NewRaftClient(conn)
	reply, err1 := c.Vote(context.Background(), request)
	if err1 != nil {
		log.Println(err1)
		return nil, err1
	}

	return reply, nil
}

func (hub MessageExchangeImpl) Append(peer string, request *pb.AppendRequest) (*pb.AppendReply, error) {
	conn, err := grpc.Dial(peer)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer conn.Close()

	c := pb.NewRaftClient(conn)
	reply, err1 := c.Append(context.Background(), request)
	if err1 != nil {
		log.Println(err)
		return nil, err1
	}

	return reply, nil
}

func (hub MessageExchangeImpl) Put(peer string, request *pb.PutRequest) (*pb.PutReply, error) {
	conn, err := grpc.Dial(peer)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer conn.Close()

	c := pb.NewRaftClient(conn)
	reply, err1 := c.Put(context.Background(), request)
	if err1 != nil {
		log.Println(err)
		return nil, err1
	}

	return reply, nil
}

func NewMessageExchange() MessageExchangeImpl {
	return MessageExchangeImpl{}
}
