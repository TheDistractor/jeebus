package network

import (
	"encoding/json"
	"flag"

	"github.com/jcw/flow"
)

func init() {
	flow.Registry["sub"] = func() flow.Circuitry { return &subCmd{} }
	flow.Registry["pub"] = func() flow.Circuitry { return &pubCmd{} }
}

type subCmd struct{ flow.Gadget }

func (g *subCmd) Run() {
	c := flow.NewCircuit()
	c.Add("s", "MQTTSub")
	c.Add("p", "Printer")
	c.Connect("s.Out", "p.In", 0)
	c.Feed("s.Port", ":1883")
	if flag.NArg() > 1 {
		c.Feed("s.Topic", flag.Arg(1))
	}
	c.Run() // never ends
}

type pubCmd struct{ flow.Gadget }

func (g *pubCmd) Run() {	
	c := flow.NewCircuit()
	c.Add("p", "MQTTPub")
	c.Feed("p.Port", ":1883")
	for i := 1; i+1 < flag.NArg(); i += 2 {
		topic := flag.Arg(i)
		payload := []byte(flag.Arg(i+1))
		var any interface{}
		if err := json.Unmarshal(payload, &any); err != nil {
			any = flag.Arg(i+1) // didn't parse as JSON, pass it in as string
		}
		c.Feed("p.In", flow.Tag{topic, any})
	}
	c.Run() // never ends
}
