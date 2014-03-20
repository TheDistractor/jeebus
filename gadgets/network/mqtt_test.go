package network

import (
	"github.com/jcw/flow"
	_ "github.com/jcw/flow/gadgets"
)

func ExampleMQTTSub() {
	// The following example requires an MQTT server running on std port 1883.
	g := flow.NewCircuit()
	g.Add("s", "MQTTSub")
	g.Feed("s.Port", ":1883")
	g.Feed("s.Topic", "#")
	g.Run()
}

func ExampleMQTTPub() {
	// The following example requires an MQTT server running on std port 1883.
	g := flow.NewCircuit()
	g.Add("p", "MQTTPub")
	g.Feed("p.Port", ":1883")
	g.Feed("p.In", flow.Tag{"Hello", "world"})
	g.Run()
}
