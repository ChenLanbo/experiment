package myraft

import "flag"
import "time"

import mynet "github.com/chenlanbo/experiment/myraft/net"

var ChannelWaitTimeout = time.Millisecond * time.Duration(*flag.Int64("channel_wait_timeout", 3000, "channel wait timeout"))

type NodeMaster struct {
    me          int
    peers       []string
    peer_server *mynet.PeerServer
}

func (s *NodeMaster) Start() {
}

func (s *NodeMaster) Shutdown() {
}

func (s *NodeMaster) Addr() string {
    return s.peers[s.me]
}

func NewNodeMaster(me int, peers []string) *NodeMaster {
    server := &NodeMaster{}
    server.me = me
    server.peers = peers
    // server.peer_server = mynet.NewHttpPeerServer(peers[me])

    return server
}
