package main
import (
	"flag"
	"github.com/chenlanbo/experiment/paxos/raft"
	"time"
	"strings"
	"github.com/chenlanbo/experiment/paxos/raft/bin"
)

func main() {
	flag.Parse()

	server := raft.NewRaftServer(strings.Split(*bin.Peers, ","), *bin.Me)
	time.Sleep(time.Second)
	server.Start()

	for !server.Stopped() {
		time.Sleep(time.Second * 10)
	}
}
