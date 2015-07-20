package raft
import (
	"net"
	"log"
	"google.golang.org/grpc"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"time"
	"golang.org/x/net/context"
	"github.com/golang/protobuf/proto"
	"testing"
)

var (
	endpoint = "localhost:50000"
)

func TestRaftServer(t *testing.T) {
	l, err1 := net.Listen("tcp", endpoint)
	if err1 != nil {
		log.Fatal(err1)
		return
	}

	s := grpc.NewServer()
	pb.RegisterRaftServerServer(s, &RaftServer{})
	go func() {
		s.Serve(l)
	} ()

	<- time.After(time.Second)

	c, err2 := grpc.Dial(endpoint)
	if err2 != nil {
		return
	}
	defer c.Close()

	cli := pb.NewRaftServerClient(c)

	request := &pb.PutRequest{
		Key:proto.String("abc"),
		Value:make([]byte, 8)}

	reply, _:= cli.Put(context.Background(), request)
	if reply != nil {
		t.Log(*reply)
	}
}
