package rpc

import (
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"google.golang.org/grpc"
	"log"
	"golang.org/x/net/context"
	"time"
)

type MessageExchange interface {

	Vote(peer string, request *pb.VoteRequest) (*pb.VoteReply, error)

	Append(peer string, request *pb.AppendRequest) (*pb.AppendReply, error)

	Put(peer string, request *pb.PutRequest) (*pb.PutReply, error)
}

type MessageExchangeImpl struct {}

func (hub MessageExchangeImpl) Vote(peer string, request *pb.VoteRequest) (*pb.VoteReply, error) {
	conn, err := grpc.Dial(peer, grpc.WithTimeout(time.Second * 2))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer conn.Close()

	c := pb.NewRaftClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond * 500)
	defer cancel()
	reply, err1 := c.Vote(ctx, request)
	if err1 != nil {
		log.Println(err1)
		return nil, err1
	}

	return reply, nil
}

func (hub MessageExchangeImpl) Append(peer string, request *pb.AppendRequest) (*pb.AppendReply, error) {
	conn, err := grpc.Dial(peer, grpc.WithTimeout(time.Second * 2))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer conn.Close()

	c := pb.NewRaftClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond * 500)
	defer cancel()
	reply, err1 := c.Append(ctx, request)
	if err1 != nil {
		log.Println(err1)
		return nil, err1
	}

	return reply, nil
}

func (hub MessageExchangeImpl) Put(peer string, request *pb.PutRequest) (*pb.PutReply, error) {
	conn, err := grpc.Dial(peer, grpc.WithTimeout(time.Second * 2))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer conn.Close()

	c := pb.NewRaftClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond * 500)
	defer cancel()
	reply, err1 := c.Put(ctx, request)
	if err1 != nil {
		log.Println(err1)
		return nil, err1
	}

	return reply, nil
}

func NewMessageExchange() MessageExchangeImpl {
	return MessageExchangeImpl{}
}
