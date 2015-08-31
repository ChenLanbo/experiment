package main
import (
	"flag"
	"github.com/chenlanbo/experiment/paxos/raft/rpc"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"math/rand"
	"os"
	"strings"
	"github.com/chenlanbo/experiment/paxos/raft/bin"
)

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		println("Argument length is not 1")
		os.Exit(1)
	}

	if flag.Args()[0] == "put" {
		sendPutToPeers(strings.Split(*bin.Peers, ","), *bin.Key, []byte(*bin.Value))
	} else if flag.Args()[0] == "get" {
		sendGetToPeers(strings.Split(*bin.Peers, ","), *bin.Key)
	}
}

func sendPutToPeers(peers []string, key string, value []byte) {
	exchange := rpc.NewMessageExchange()
	request := &pb.PutRequest{
		Key:proto.String(key),
		Value:value}
	tries := 0
	to := peers[0]

	for ; tries != 5; tries++ {
		reply, err := exchange.Put(to, request)
		if err != nil {
			continue
		}
		if reply.GetSuccess() {
			println("Put succeeded")
			break
		} else {
			to = reply.GetLeaderId()
			if to == "" {
				to = peers[rand.Intn(len(peers))]
			}
		}
	}

	if tries == 5 {
		os.Exit(1)
	}
}

func sendGetToPeers(peers []string, key string) {
	exchange := rpc.NewMessageExchange()
	request := &pb.GetRequest{
		Key:proto.String(key)}
	tries := 0
	to := peers[0]

	for ; tries < 5; tries++ {
		reply, err := exchange.Get(to, request)
		if err != nil {
			continue
		}
		if reply.GetSuccess() {
			println("Get value:", string(reply.GetValue()))
			break
		} else {
			to = reply.GetLeaderId()
			if to == "" {
				to = peers[rand.Intn(len(peers))]
			}
		}
	}

	if tries == 5 {
		os.Exit(1)
	}
}