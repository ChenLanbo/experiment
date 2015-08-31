package bin

import "flag"

var (
	Peers = flag.String("peers", "", "")
	Me = flag.Int("me", -1, "")
	Key = flag.String("key", "", "")
	Value = flag.String("value", "", "")
)
