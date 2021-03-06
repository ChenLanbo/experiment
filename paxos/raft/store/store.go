package store

import (
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"sync"
	"errors"
	"flag"
	"time"
)

var (
	batchSize = flag.Int("raft_store_batch_size", 10, "")
	pollTimeout = flag.Int("raft_store_poll_timeout", 10, "")
)

type Store struct {
	currentTerm uint64
	commitIndex uint64
	lastApplied uint64
	logs []pb.Log
	snapshot map[string][]byte

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

func (s *Store) SetCommitIndex(commitIndex uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.logs[len(s.logs) - 1].GetLogId() < commitIndex {
		return
	}

	if s.commitIndex < commitIndex {
		// 1. Apply new committed logs to snapshot
		for i := s.commitIndex + 1; i <= commitIndex; i++ {
			logData := s.logs[i].GetData()
			if logData != nil && logData.GetKey() != "" {
				s.snapshot[logData.GetKey()] = logData.GetValue()
			}
		}
		// 2. Update commit index
		s.commitIndex = commitIndex
	}
}

func (s *Store) LatestIndex() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.logs[len(s.logs) - 1].GetLogId()
}

// Check if another peer's log is at least up to date as my log:
//
// Returns:
//   true if otherLogTerm > myTerm or if otherLogTerm == myTerm and otherLogId >= myLogId
//   false if otherLogTerm < myTerm or if otherLogTerm == myTerm and otherLogId < myLogId
func (s *Store) OtherLogUpToDate(otherLogId, otherLogTerm uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	myLogId := s.logs[len(s.logs) - 1].GetLogId()
	myTerm := s.logs[len(s.logs) - 1].GetTerm()

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

	if logId >= uint64(len(s.logs)) || term != s.logs[logId].GetTerm() {
		return false
	}

	return true
}

func (s *Store) Read(logId uint64) *pb.Log {
	if logId >= uint64(len(s.logs)) {
		return nil
	}

	return &s.logs[logId]
}

func (s *Store) Poll(logId uint64) *pb.Log {
	l := s.Read(logId)
	if l != nil {
		return l
	}
	<- time.After(time.Millisecond * time.Duration(*pollTimeout))
	return s.Read(logId)
}

func (s *Store) Append(commitIndex uint64, logs []pb.Log) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check log continuous
	var prevLogId uint64 = 0
	for _, l := range(logs) {
		if prevLogId == 0 {
			prevLogId = l.GetLogId()
		} else {
			if prevLogId + 1 != l.GetLogId() {
				return errors.New("LogId not continuous")
			} else {
				prevLogId = l.GetLogId()
			}
		}
	}

	if len(logs) > 0 && logs[0].GetLogId() > s.logs[len(s.logs) - 1].GetLogId() + 1 {
		// Gap
		return errors.New("LogId not continuous")
	}

	latestLogId := s.logs[len(s.logs) - 1].GetLogId()
	if len(logs) > 0 && latestLogId < logs[len(logs) - 1].GetLogId() {
		latestLogId = logs[len(logs) - 1].GetLogId()
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
	if s.commitIndex < commitIndex {
		for i := s.commitIndex + 1; i <= commitIndex; i++ {
			logData := s.logs[i].GetData()
			if logData != nil && logData.GetKey() != "" {
				s.snapshot[logData.GetKey()] = logData.GetValue()
			}
		}
		s.commitIndex = commitIndex
	}

	return nil
}

func (s *Store) Write(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newLog := pb.Log{}
	newLog.LogId = proto.Uint64(s.logs[len(s.logs) - 1].GetLogId() + 1)
	newLog.Term = proto.Uint64(s.currentTerm)

	s.logs = append(s.logs, newLog)
}

func (s *Store) WriteKeyValue(key string, value []byte) uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	newLog := pb.Log{}
	newLog.LogId = proto.Uint64(s.logs[len(s.logs) - 1].GetLogId() + 1)
	newLog.Term = proto.Uint64(s.currentTerm)
	newLog.Data = &pb.Log_Data{
		Key:proto.String(key),
		Value:value}

	s.logs = append(s.logs, newLog)
	return newLog.GetLogId()
}

func (s *Store) GetKeyValue(key string) []byte {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, present := s.snapshot[key]
	if !present {
		return nil
	}

	value := make([]byte, len(s.snapshot[key]))
	copy(value, s.snapshot[key])
	return value
}

func NewStore() (*Store) {
	s := &Store{}

	s.currentTerm = 0
	s.commitIndex = 0
	s.lastApplied = 0
	s.logs = make([]pb.Log, 1)
	s.logs[0] = pb.Log{Term:proto.Uint64(0),
	                   LogId:proto.Uint64(0)}
	s.snapshot = make(map[string][]byte)

	return s
}