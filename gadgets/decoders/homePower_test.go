package decoders

import (
	"github.com/jcw/flow"
)

func ExampleHomePower() {
	g := flow.NewCircuit()
	g.Add("d", "Node-homePower")
	g.Feed("d.In", []byte{9, 213, 11, 68, 235, 151, 90, 99, 6, 88, 198, 136, 89})
	g.Feed("d.In", []byte{9, 213, 11, 68, 235, 153, 90, 84, 6, 88, 198, 136, 89})
	g.Run()
	// Output:
	// Lost map[string]int: map[<reading>:1 c1:3029 p1:78 c2:23191 p2:11009 c3:50776 p3:785]
	// Lost map[string]int: map[<reading>:1 c2:23193 p2:11111]
}
