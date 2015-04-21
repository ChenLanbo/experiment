package net

import "bytes"
import "io/ioutil"
import "log"
import "net/http"
import "net/url"

import myproto "github.com/chenlanbo/experiment/myraft/proto"
import "github.com/golang/protobuf/proto"

type HttpPeer struct {
    addr string
    client *http.Client
}

func NewHttpPeer(address string) *HttpPeer {
    peer := &HttpPeer{}
    peer.addr = address
    peer.client = &http.Client{}

    return peer
}

func GetVoteRequestURL(addr string) string {
    url := &url.URL{
        Scheme: "http",
        Host: addr,
        Path: "vote",
    }
    log.Println("URL:", url)
    return url.String()
}

func (peer *HttpPeer) SendVoteRequest(vote_req *myproto.VoteRequest) (*myproto.VoteReply, error) {
    raw_vote_req, err := proto.Marshal(vote_req)
    if err != nil {
        log.Fatal("protobuf marshaling error:", err)
        return nil, err
    }

    var req *http.Request = nil
    req, err = http.NewRequest("post", GetVoteRequestURL(peer.addr), bytes.NewReader(raw_vote_req))
    if err != nil {
        log.Println("error new http request:", err)
        return nil, err
    }

    // call peer
    var rep *http.Response = nil
    rep, err = peer.client.Do(req)
    if err != nil {
        log.Println("error while getting response:", err)
        return nil, err
    }

    raw_vote_rep := make([]byte, 0)
    raw_vote_rep, err = ioutil.ReadAll(rep.Body)
    if err != nil {
        log.Println("error while reading body:", err)
        return nil, err
    }

    vote_rep := &myproto.VoteReply{}
    err = proto.Unmarshal(raw_vote_rep, vote_rep)
    if err != nil {
        log.Fatal("protobuf unmarshaling error:", err)
        return nil, err
    }

    return vote_rep, nil
}
