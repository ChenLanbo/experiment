package store
import (
	"testing"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
	"reflect"
)

func TestAppend(t *testing.T) {
	store := NewStore()

	batch1 := make([]pb.Log, 5)
	for i := 0; i < 5; i++ {
		batch1[i] = pb.Log{LogId:proto.Uint64(uint64(i + 1)), Term:proto.Uint64(1)}
	}

	err := store.Append(0, batch1)
	if err != nil && store.LatestIndex() != 5 && store.CommitIndex() != 0 {
		t.Fatal("error")
		t.Fail()
	}

	err = store.Append(3, make([]pb.Log, 0))
	if err != nil && store.LatestIndex() != 5 && store.CommitIndex() != 3 {
		t.Fatal("error")
		t.Fail()
	}

	batch2 := make([]pb.Log, 5)
	for i := 0; i < 5; i++ {
		batch2[i] = pb.Log{LogId:proto.Uint64(uint64(6 + i)), Term:proto.Uint64(2)}
	}

	err = store.Append(6, batch2)
	if err != nil && store.LatestIndex() != 10 && store.CommitIndex() != 6 {
		t.Fatal("error ", err, store.LatestIndex(), store.CommitIndex())
		t.Fail()
	}

	batch3 := make([]pb.Log, 1)
	batch3[0] = pb.Log{LogId:proto.Uint64(100), Term:proto.Uint64(3)}

	err = store.Append(7, batch3)
	if err == nil {
		t.Fatal("error")
		t.Fail()
	}

	err = store.Append(100, make([]pb.Log, 0))
	if err == nil {
		t.Fatal("error")
		t.Fail()
	}
}

func TestPoll(t *testing.T) {
	*pollTimeout = 2000
	store := NewStore()

	batch1 := make([]pb.Log, 5)
	for i := 0; i < 5; i++ {
		batch1[i] = pb.Log{LogId:proto.Uint64(uint64(i + 1)), Term:proto.Uint64(1)}
	}

	err := store.Append(0, batch1)
	if err != nil && store.LatestIndex() != 5 && store.CommitIndex() != 0 {
		t.Fatal("error")
		t.Fail()
	}

	l := store.Poll(1)
	if l == nil || *l.LogId != 1 || *l.Term != 1{
		t.Fatal("error")
		t.Fail()
	}

	l = store.Poll(6)
	if l != nil {
		t.Fatal("error")
		t.Fail()
	}
}

func TestCommit(t *testing.T) {
	store := NewStore()

	store.IncrementCurrentTerm()
	store.WriteKeyValue("abc", []byte("abc"))
	store.SetCommitIndex(store.LatestIndex())

	value := store.GetKeyValue("abc")
	if value == nil || !reflect.DeepEqual([]byte("abc"), value) {
		t.Fail()
	}

	logs := make([]pb.Log, 0)
	logs = append(
		logs,
		pb.Log{
			Term:proto.Uint64(1),
			LogId:proto.Uint64(2),
			Data:&pb.Log_Data{
				Key:proto.String("key1"),
				Value:[]byte("value1")}})
	logs = append(
		logs,
		pb.Log{
			Term:proto.Uint64(1),
			LogId:proto.Uint64(3),
			Data:&pb.Log_Data{
				Key:proto.String("key2"),
				Value:[]byte("value2")}})
	store.Append(1, logs)
	store.Append(3, make([]pb.Log, 0))

	value = store.GetKeyValue("key1")
	if value == nil || !reflect.DeepEqual([]byte("value1"), value) {
		t.Fail()
	}

	value = store.GetKeyValue("key2")
	if value == nil || !reflect.DeepEqual([]byte("value2"), value) {
		t.Fail()
	}
}