package net

import "errors"
import "io/ioutil"
import "log"
import "net"
import "net/http"
import "sync"

import myproto "github.com/chenlanbo/experiment/myraft/proto"
import "github.com/golang/protobuf/proto"

type VoteHandler struct {
    server *HttpPeerServer
}

func (h *VoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    log.Println("Server gets new vote request")
    raw, err := ioutil.ReadAll(r.Body)
    if err != nil {
        // ignore
        return
    }

    req := &myproto.VoteRequest{}
    err = proto.Unmarshal(raw, req)
    if err != nil {
        // ignore
        return
    }

    cb := &VoteCallback {
        Req: req,
        Rep: make(chan myproto.VoteReply, 1),
    }

    h.server.mu.Lock()
    h.server.voteChan <- cb
    h.server.mu.Unlock()

    rep := <-cb.Rep
    log.Println("Server sends response", rep)
    raw, err = proto.Marshal(&rep)
    if err != nil {
        // ignore
        return
    }
    w.Write(raw)
}

type AppendHandler struct {
    server *HttpPeerServer
}

func (h *AppendHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    log.Println("New append request")
    raw, err := ioutil.ReadAll(r.Body)
    if err != nil {
        // ignore
        return
    }

    req := &myproto.AppendRequest{}
    err = proto.Unmarshal(raw, req)
    if err != nil {
        // ignore
        return
    }

    cb := &AppendCallback {
        Req: req,
        Rep: make(chan myproto.AppendReply, 1),
    }

    h.server.mu.Lock()
    h.server.appendChan <- cb
    h.server.mu.Unlock()

    rep := <-cb.Rep
    raw, err = proto.Marshal(&rep)
    if err != nil {
        // ignore
        return
    }
    w.Write(raw)
}

type HttpPeerServer struct {
    mu sync.Mutex
    ln *net.TCPListener
    addr string
    voteChan chan *VoteCallback
    appendChan chan *AppendCallback
}

func NewHttpPeerServer(addr string) *HttpPeerServer {
    server := &HttpPeerServer{}
    server.mu = sync.Mutex{}
    server.addr = addr
    server.voteChan = make(chan *VoteCallback, 1)
    server.appendChan = make(chan *AppendCallback, 1)

    ln, err := net.Listen("tcp", addr)
    if err != nil {
        log.Fatal(err)
    }
    var ok bool
    server.ln, ok = ln.(*net.TCPListener)
    if !ok {
        log.Fatal(errors.New("Invalide tcp listener"))
    }

    return server
}

func (server *HttpPeerServer) Start() {
    mux := http.NewServeMux();
    mux.Handle("/vote", &VoteHandler{server})
    mux.Handle("/append", &AppendHandler{server})

    s := &http.Server {
        Handler: mux,
    }
    go func() {
        err := s.Serve(server.ln)
        if err != nil {
            log.Println(err)
        }
    } ()
}

func (server *HttpPeerServer) Shutdown() {
    err := server.ln.Close()
    if err != nil {
        log.Println("Error shutting down server:", err)
    }
}

func (server *HttpPeerServer) VoteChan() chan *VoteCallback {
    return server.voteChan
}

func (server *HttpPeerServer) AppendChan() chan *AppendCallback {
    return server.appendChan
}
