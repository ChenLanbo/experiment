package main

//import "net"
//import "net/rpc"

//import "github.com/chenlanbo/experiment/paxos"

func main() {
  m := make(map[int]int)
  m[1] = 1
  m[2] = 2

  for k, _ := range m {
    delete(m, k)
  }

  println(len(m))
}
