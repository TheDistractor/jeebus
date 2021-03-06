// The (very early) start of an FBP parser.
package fbpparse

import (
	"strings"

	"github.com/jcw/flow"
)

func init() {
	flow.Registry["FbpParser"] = func() flow.Circuitry { return new(FbpParser) }
}

// FbpParser processes a graph definition in FBP syntax.
type FbpParser struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
}

// Start collecting lines and parse the resulting string.
func (w *FbpParser) Run() {
	lines := []string{}
	for m := range w.In {
		if tag, ok := m.(flow.Tag); ok {
			switch tag.Tag {
			case "<close>":
				w.parseFbp(lines)
				fallthrough
			case "<open>":
				lines = []string{}
			}
		} else if s, ok := m.(string); ok {
			lines = append(lines, s)
		}
	}
	w.parseFbp(lines)
}

func (w *FbpParser) parseFbp(lines []string) {
	if len(lines) > 0 {
		fbp := &Fbp{Buffer: strings.Join(lines, "\n")}
		fbp.Init()
		err := fbp.Parse()
		flow.Check(err)
		// fbp.Execute()
		w.Out.Send(true)
		// fbp.PrintSyntaxTree()
	}
}
