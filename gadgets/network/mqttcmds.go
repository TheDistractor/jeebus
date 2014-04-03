package network

import (
	"github.com/jcw/flow"
)

func init() {
	flow.Registry["mqttsub"] = func() flow.Circuitry { return new(subCmd) }
	flow.Registry["mqttpub"] = func() flow.Circuitry { return new(pubCmd) }
}

type subCmd struct{ flow.Gadget }

func (g *subCmd) Run() {
	c := flow.NewCircuit()
	c.Add("c", "CmdLine")
	c.Add("s", "MQTTSub")
	c.Add("p", "Printer")
	c.Connect("c.Out", "s.In", 0)
	c.Connect("s.Out", "p.In", 0)
	c.Feed("a.Type", "skip")
	c.Feed("s.Port", ":1883")
	c.Run() // never ends
}

type pubCmd struct{ flow.Gadget }

func (g *pubCmd) Run() {
	c := flow.NewCircuit()
	c.Add("c", "CmdLine")
	c.Add("p", "MQTTPub")
	c.Connect("a.Out", "p.In", 0)
	c.Feed("c.Type", "skip,tags,json")
	c.Feed("p.Port", ":1883")
	c.Run()
}
