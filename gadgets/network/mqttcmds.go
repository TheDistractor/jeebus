package network

import (
	"github.com/jcw/flow"
)

func init() {
	flow.Registry["see"] = func() flow.Circuitry { return &seeCmd{} }
}

type seeCmd struct{ flow.Gadget }

func (g *seeCmd) Run() {
	c := flow.NewCircuit()
	c.Add("s", "MQTTSub")
	c.Add("p", "Printer")
	c.Connect("s.Out", "p.In", 0)
	c.Feed("s.Port", ":1883")
	c.Feed("s.Topic", "#")
	c.Run() // never ends
}
