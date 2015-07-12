package main

import (
	"net"
	"log"
	"golang.org/x/net/context"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

type server struct{}

func (s *server) PingPong(ctx context.Context, ping *pb.Ping) (*pb.Pong, error) {
	return &pb.Pong{Message:proto.String("abc")}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":8080")

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterPingServerServer(s, &server{})
	s.Serve(lis)
}
