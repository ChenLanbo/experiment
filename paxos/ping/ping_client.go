package main
import (
	"google.golang.org/grpc"
	"log"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"golang.org/x/net/context"
	"github.com/golang/protobuf/proto"
)

func main() {
	conn, err := grpc.Dial("localhost:8080")
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()
	c := pb.NewPingServerClient(conn)

	for i := 0; i < 10; i++ {
		r, err := c.PingPong(context.Background(), &pb.Ping{Message:proto.String("abc")})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Greeting: %s\n", *r.Message)
	}
}