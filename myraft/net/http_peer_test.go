package net

import "log"
import "os"
import "testing"
import "time"

import myproto "github.com/chenlanbo/experiment/myraft/proto"
import "github.com/golang/protobuf/proto"

const ADDR string = "127.0.0.1:8080"

var server *HttpPeerServer

func TestSendVoteRequest(t *testing.T) {
    expected_rep := myproto.VoteReply {
        CurrentTerm: proto.Uint64(2),
        Granted: proto.Bool(false),
    }
    go func() {
        cb := <-server.VoteChan
        cb.Rep <- expected_rep
    } ()

    peer := NewHttpPeer(ADDR)
    req := &myproto.VoteRequest {
        Term: proto.Uint64(1),
        CandidateId: proto.String(ADDR),
        LastLogIndex: proto.Uint64(2),
        LastLogTerm: proto.Uint64(1),
    }

    rep, err := peer.VoteRequest(req)
    log.Println("Got reply:", rep)
    if err != nil {
        t.Fail()
    }
}

func TestMain(m *testing.M) {
    server = NewHttpPeerServer(ADDR)
    server.Start()
    log.Println("Server started")
    time.Sleep(time.Second)
    ret := m.Run()
    server.Shutdown()
    os.Exit(ret)
}

