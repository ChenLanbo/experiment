package store

import (
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"sync"
	"errors"
	"log"
)

type Store struct {
	currentTerm uint64
	commitIndex uint64
	lastApplied uint64
	logs []pb.Log

	mu sync.Mutex
}

func (s *Store) CurrentTerm() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.currentTerm
}

func (s *Store) IncrementCurrentTerm() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.currentTerm++
}

func (s *Store) SetCurrentTerm(term uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if term > s.currentTerm {
		s.currentTerm = term
	}
}

func (s *Store) CommitIndex() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.commitIndex
}

func (s *Store) LatestIndex() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Print("Log length: ", len(s.logs))

	return *(s.logs[len(s.logs) - 1]).LogId
}

func (s *Store) OtherLogUpToDate(otherLogId, otherLogTerm uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	myLogId := *s.logs[len(s.logs) - 1].LogId
	myTerm := *s.logs[len(s.logs) - 1].Term

	if myTerm < otherLogTerm {
		return true
	} else if myTerm > otherLogTerm {
		return false
	} else {
		return otherLogId >= myLogId
	}
}

func (s *Store) Match(logId, term uint64) (bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if logId >= uint64(len(s.logs)) || term != *s.logs[logId].Term  {
		return false
	}

	return true
}

func (s *Store) Read(logId uint64) *pb.Log {
	s.mu.Lock()
	defer s.mu.Unlock()

	if logId >= uint64(len(s.logs)) {
		return nil
	}

	return &s.logs[logId]
}

func (s *Store) Append(commitIndex uint64, logs []pb.Log) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check log continuous
	var prevLogId uint64 = 0
	for _, l := range(logs) {
		if prevLogId == 0 {
			prevLogId = *l.LogId
		} else {
			if prevLogId + 1 != *l.LogId {
				return errors.New("LogId not continuous")
			} else {
				prevLogId = *l.LogId
			}
		}
	}

	if len(logs) > 0 && *logs[0].LogId > *s.logs[len(s.logs) - 1].LogId + 1 {
		return errors.New("LogId not continuous")
	}

	latestLogId := *s.logs[len(s.logs) - 1].LogId
	if len(logs) > 0 && latestLogId < *logs[len(logs) - 1].LogId {
		latestLogId = *logs[len(logs) - 1].LogId
	}
	if commitIndex > latestLogId {
		return errors.New("Commit index out of bound")
	}

	// Append logs
	for _, l := range(logs) {
		// log.Print("Append log: ", *l.LogId)
		if *l.LogId < uint64(len(s.logs)) {
			s.logs[*l.LogId] = l
		} else {
			s.logs = append(s.logs, l)
		}
	}

	// Update commit index
	if commitIndex > s.commitIndex {
		s.commitIndex = commitIndex
	}

	return nil
}

func NewStore() (*Store) {
	s := &Store{}

	s.currentTerm = 0
	s.commitIndex = 0
	s.lastApplied = 0
	s.logs = make([]pb.Log, 1)
	s.logs[0] = pb.Log{Term:proto.Uint64(0),
	                   LogId:proto.Uint64(0)}

	return s
}