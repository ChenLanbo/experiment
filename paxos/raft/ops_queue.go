package raft

import (
	"time"
	"flag"
)

var (
	queueSize = flag.Uint("raft_ops_queue_size", 10, "size of operation queue")
	queuePullTimeout = flag.Uint("raft_ops_queue_pull_timeout", 1000, "pull timeout")
)

type OperationQueue struct {
	queue chan *RaftOperation
}

func (q *OperationQueue) Push(op *RaftOperation) {
	if op == nil {
		return
	}

	q.queue <- op
}

func (q *OperationQueue) Pull(timeout time.Duration) (*RaftOperation) {
	select {
	case op := <- q.queue:
		return op
	case <- time.After(timeout):
		return nil
	}
}

func (q *OperationQueue) Close() {
	close(q.queue)
}

func NewOperationQueue() (*OperationQueue) {
	return &OperationQueue{queue:make(chan *RaftOperation, *queueSize)}
}