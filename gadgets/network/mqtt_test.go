package network

import (
	"github.com/jcw/flow"
	_ "github.com/jcw/flow/gadgets"
)

func ExampleMQTTSub() {
	// The following example never ends.
	g := flow.NewGroup()
	g.Add("s", "MQTTSub")
	g.Set("s.Port", ":1883")
	g.Set("s.Topic", "#")
	g.Run()
}

func ExampleMQTTPub() {
	// The following example never ends.
	g := flow.NewGroup()
	g.Add("p", "MQTTPub")
	g.Set("p.Port", ":1883")
	g.Set("p.In", []string{"Hello", "world"})
	g.Run()
}
