package rfdata

import (
	"github.com/jcw/flow"
)

func ExampleRF12demo() {
	g := flow.NewCircuit()
	g.Add("rf", "Sketch-RF12demo")
	g.Feed("rf.In", "[RF12demo.12] _ i31* g5 @ 868 MHz c1 q1")
	g.Feed("rf.In", "OK 9 187 176 69 235 249 6 192 234 6 74 190 18 (-66)")
	g.Feed("rf.In", "OK 37 2 107 185 0 (-76)")
	g.Feed("rf.In", "OK 197 (-60)")
	g.Run()
	// Output:
	// Lost map[string]int: map[<RF12demo>:12 band:868 group:5 id:31]
	// Lost string: [RF12demo.12] _ i31* g5 @ 868 MHz c1 q1
	// Lost map[string]int: map[<node>:9 rssi:-66]
	// Lost []uint8: [9 187 176 69 235 249 6 192 234 6 74 190 18]
	// Lost map[string]int: map[<node>:5 rssi:-76]
	// Lost []uint8: [37 2 107 185 0]
	// Lost map[string]int: map[<node>:5 rssi:-60]
	// Lost []uint8: [197]
}

func ExampleNodeMap() {
	g := flow.NewCircuit()
	g.Add("nm", "NodeMap")
	g.Feed("nm.Info", "RFg5i4,roomNode,kitchen")
	g.Feed("nm.In", map[string]int{"<RF12demo>": 1, "group": 5})
	g.Feed("nm.In", map[string]int{"<node>": 3})
	g.Feed("nm.In", map[string]int{"<node>": 4})
	g.Feed("nm.In", map[string]int{"<node>": 5})
	g.Run()
	// Output:
	// Lost map[string]int: map[<RF12demo>:1 group:5]
	// Lost map[string]int: map[<node>:3]
	// Lost map[string]int: map[<node>:4]
	// Lost flow.Tag: {<location> kitchen}
	// Lost flow.Tag: {<dispatch> roomNode}
	// Lost map[string]int: map[<node>:5]
}
