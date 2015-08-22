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

  x := make(chan bool, 0)
  // close(x)
  y := <- x
  println(y)
}
