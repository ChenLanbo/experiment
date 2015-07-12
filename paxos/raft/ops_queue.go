package raft

import (
	"time"
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

func NewOperationQueue() (*OperationQueue) {
	return &OperationQueue{queue:make(chan *RaftOperation, 10)}
}