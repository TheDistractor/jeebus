package decoders

import (
	"github.com/jcw/flow"
)

func init() {
	flow.Registry["Node-ookRelay"] = func() flow.Circuitry {
		g := flow.NewCircuit()
		g.AddCircuitry("or", &OokRelay{})
		g.Add("dp", "Dispatcher")
		g.Connect("or.Out", "dp.In", 0)
		g.Feed("dp.Prefix", "Node-ook")
		g.Label("In", "or.In")
		g.Label("Out", "dp.Out")
		g.Label("Rej", "dp.Rej")
		return g
	}
}

var ookDecoders = []string{
	"Dcf", "Viso", "Emx", "Ksx", "Fsx", "Orsc", "Cres", "Kaku",
	"Xrf", "Hez", "Elro", "11?", "12?", "13?", "14?", "15?",
}

// Decoder for the "ookRelay.ino" sketch. Registers as "Node-ookRelay".
type OokRelay struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
}

// Start decoding ookRelay packets.
func (w *OokRelay) Run() {
	for m := range w.In {
		if v, ok := m.([]byte); ok {
			offset := 1
			for offset < len(v) {
				typ := int(v[offset] & 0x0F)
				size := int(v[offset] >> 4)
				offset++

				// insert a new decoder request
				tag := ookDecoders[typ]
				w.Out.Send(flow.Tag{"<dispatch>", tag})
				w.Out.Send(v[offset : offset+size])

				offset += size
			}
		} else {
			w.Out.Send(m)
		}
	}
}
