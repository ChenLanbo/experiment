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

	Get(peer string, request *pb.GetRequest) (*pb.GetReply, error)
}

type MessageExchangeImpl struct {}

func (hub MessageExchangeImpl) Vote(peer string, request *pb.VoteRequest) (*pb.VoteReply, error) {
	conn, err := grpc.Dial(peer,
		grpc.WithInsecure(), grpc.WithTimeout(time.Second * 2))
	if err != nil {
		log.Println("Fail to send vote to", peer, err)
		return nil, err
	}
	defer conn.Close()

	c := pb.NewRaftClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond * 500)
	defer cancel()
	// log.Println("Send vote to", peer)
	reply, err1 := c.Vote(ctx, request)
	if err1 != nil {
		log.Println(err1)
		return nil, err1
	}

	return reply, nil
}

func (hub MessageExchangeImpl) Append(peer string, request *pb.AppendRequest) (*pb.AppendReply, error) {
	conn, err := grpc.Dial(peer,
		grpc.WithInsecure(), grpc.WithTimeout(time.Second * 2))
	if err != nil {
		log.Println("Fail to send append to", peer, err)
		return nil, err
	}
	defer conn.Close()

	c := pb.NewRaftClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond * 500)
	defer cancel()
	// log.Println("Send append to", peer)
	reply, err1 := c.Append(ctx, request)
	if err1 != nil {
		log.Println(err1)
		return nil, err1
	}

	return reply, nil
}

func (hub MessageExchangeImpl) Put(peer string, request *pb.PutRequest) (*pb.PutReply, error) {
	conn, err := grpc.Dial(peer,
		grpc.WithInsecure(), grpc.WithTimeout(time.Second * 2))
	if err != nil {
		log.Println("Fail to send put to", peer, err)
		return nil, err
	}
	defer conn.Close()

	c := pb.NewRaftClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond * 500)
	defer cancel()
	// log.Println("Send put to", peer)
	reply, err1 := c.Put(ctx, request)
	if err1 != nil {
		log.Println(err1)
		return nil, err1
	}

	return reply, nil
}

func (hub MessageExchangeImpl) Get(peer string, request *pb.GetRequest) (*pb.GetReply, error) {
	conn, err := grpc.Dial(peer,
		grpc.WithInsecure(), grpc.WithTimeout(time.Second * 2))
	if err != nil {
		log.Println("Fail to send get to", peer, err)
		return nil, err
	}
	defer conn.Close()

	c := pb.NewRaftClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond * 500)
	defer cancel()
	// log.Println("Send get to", peer)
	reply, err1 := c.Get(ctx, request)
	if err1 != nil {
		log.Println(err1)
		return nil, err1
	}

	return reply, nil
}

func NewMessageExchange() MessageExchangeImpl {
	return MessageExchangeImpl{}
}
