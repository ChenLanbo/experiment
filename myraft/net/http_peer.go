package net

import "bytes"
import "io/ioutil"
import "log"
import "net/http"
import "net/url"

import myproto "github.com/chenlanbo/experiment/myraft/proto"
import "github.com/golang/protobuf/proto"

type HttpPeerClient struct {
    addr string
    client *http.Client
}

func NewHttpPeerClient(address string) *HttpPeerClient {
    peer := &HttpPeerClient{}
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
    // log.Println("URL:", url)
    return url.String()
}

func GetAppendRequestURL(addr string) string {
    url := &url.URL{
        Scheme: "http",
        Host: addr,
        Path: "append",
    }
    // log.Println("URL:", url)
    return url.String()
}

func (peer *HttpPeerClient) VoteRequest(vote_req *myproto.VoteRequest) (*myproto.VoteReply, error) {
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

func (peer *HttpPeerClient) AppendRequest(append_req *myproto.AppendRequest) (*myproto.AppendReply, error) {
    raw_append_req, err := proto.Marshal(append_req)
    if err != nil {
        log.Fatal("protobuf marshaling error:", err)
        return nil, err
    }

    var req *http.Request = nil;
    req, err = http.NewRequest("post", GetAppendRequestURL(peer.addr), bytes.NewReader(raw_append_req))
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

    raw_append_rep := make([]byte, 0)
    raw_append_rep, err = ioutil.ReadAll(rep.Body)
    if err != nil {
        log.Println("error while reading body:", err)
        return nil, err
    }

    append_rep := &myproto.AppendReply{}
    err = proto.Unmarshal(raw_append_rep, append_rep)
    if err != nil {
        log.Fatal("protobuf unmarshaling error:", err)
        return nil, err
    }

    return append_rep, nil
}

