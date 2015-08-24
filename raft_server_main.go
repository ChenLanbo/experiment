package main

import "golang.org/x/net/context"
import "time"
//import "net/rpc"

//import "github.com/chenlanbo/experiment/paxos"

func XX() {
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()

  go func() {
    time.Sleep(time.Second * 2)
    print("Go routine done")
    cancel()
  } ()

  select {
  case <-ctx.Done():
    print("Child done")
    print(ctx.Err());
  }

}

func main() {
  x := int64(time.Second)
  println(x)

  mp := make(map[string][]byte)
  mp["abc"] = []byte("abc")

  println(string(mp["abc"]))
}
