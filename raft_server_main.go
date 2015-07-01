package main

import "log"
import "net"
import "net/rpc"

import "github.com/chenlanbo/experiment/paxos"

func main() {
    raft_server := new(paxos.RaftServer)
    rpc.Register(raft_server)
    l, e := net.Listen("tcp", ":1234")
    if e != nil {
        log.Fatal("listen error: ", e)
    }

    for {
        if conn, err := l.Accept(); err != nil {
            log.Fatal("accept error: ", err)
        } else {
            log.Printf("new connection established\n")
            go rpc.ServeConn(conn)
        }
    }
}
