package net

import "net/http"
import "testing"

import myproto "github.com/chenlanbo/experiment/myraft/proto"
import "github.com/golang/protobuf/proto"

type testHandler struct {
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    vote_rep := &myproto.VoteReply {
        CurrentTerm: proto.Uint64(2),
        Granted: proto.Bool(false),
    }

    raw_vote_rep, _ := proto.Marshal(vote_rep)
    w.Write(raw_vote_rep)
    w.WriteHeader(200)
}

func TestSendVoteRequest(t *testing.T) {
    const addr string = ":8080"

    s := &http.Server {
        Addr: addr,
        Handler: &testHandler{},
    }
    go func() {
        s.ListenAndServe()
    }()

    peer := NewHttpPeer("127.0.0.1:8080")

    vote_req := &myproto.VoteRequest {
        Term: proto.Uint64(1),
        CandidateId: proto.String("myid"),
        LastLogIndex: proto.Uint64(2),
        LastLogTerm: proto.Uint64(1),
    }

    vote_rep, err := peer.SendVoteRequest(vote_req)

    if err != nil {
        t.Fail()
    }
    if vote_rep.GetCurrentTerm() != 2 || vote_rep.GetGranted() {
        t.Fail()
    }
    t.Log("get response:", vote_rep)
}

