// Embedded JavaScript engine.
package javascript

import (
	"github.com/golang/glog"
	"github.com/jcw/flow"
	"github.com/robertkrimen/otto"
)

func init() {
	flow.Registry["JavaScript"] = func() flow.Circuitry { return new(JavaScript) }
}

// JavaScript engine, using the github.com/robertkrimen/otto package.
type JavaScript struct {
	flow.Gadget
	In  flow.Input
	Cmd flow.Input
	Out flow.Output
}

// Start running the JavaScript engine.
func (w *JavaScript) Run() {
	if cmd, ok := <-w.Cmd; ok {
		// initial setup
		engine := otto.New()

		// define a callback for sending messages to Out
		engine.Set("emitOut", func(call otto.FunctionCall) otto.Value {
			out, err := call.Argument(0).Export()
			flow.Check(err)
			w.Out.Send(out)
			return otto.UndefinedValue()
		})

		// process the command input
		if _, err := engine.Run(cmd.(string)); err != nil {
			glog.Fatal(err)
		}

		// only start the processing loop if the "onIn" handler exists
		value, err := engine.Get("onIn")
		if err == nil && value.IsFunction() {
			for in := range w.In {
				engine.Call("onIn", nil, in)
			}
		}
	}
}
