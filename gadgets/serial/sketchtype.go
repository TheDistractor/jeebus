// +build linux darwin windows

package serial

import (
	"strings"

	"github.com/jcw/flow"
)

func init() {
	flow.Registry["SketchType"] = func() flow.Circuitry { return new(SketchType) }
}



// SketchType looks for lines of the form "[name...]" in the input stream.
// These then cause a corresponding .Feed( to be loaded dynamically.
// Registers as "SketchType".
type SketchType struct {
flow.Gadget
In  flow.Input
Out flow.Output
}

// Start transforming the "[name...]" markers in the input stream.
func (w *SketchType) Run() {
	for m := range w.In {
		if s, ok := m.(string); ok {
			if strings.HasPrefix(s, "[") && strings.Contains(s, "]") {
				tag := s[1:strings.IndexAny(s, ".]")]
				w.Out.Send(flow.Tag{"<dispatch>", tag})
		}
	}
	w.Out.Send(m)
}
}
