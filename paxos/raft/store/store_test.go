package store
import (
	"testing"
	pb "github.com/chenlanbo/experiment/paxos/protos"
	"github.com/golang/protobuf/proto"
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