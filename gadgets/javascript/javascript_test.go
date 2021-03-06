package javascript

import (
	"github.com/jcw/flow"
	_ "github.com/jcw/flow/gadgets"
)

func ExampleJavaScript() {
	g := flow.NewCircuit()
	g.Add("js", "JavaScript")
	g.Feed("js.Cmd", `console.log("Hello from Otto!");`)
	g.Run()
	// Output:
	// Hello from Otto!
}

func ExampleJavaScript_2() {
	g := flow.NewCircuit()
	g.Add("js", "JavaScript")
	g.Feed("js.Cmd", `
	  console.log("Howdy from Otto!");
	  function onIn(v) {
	    console.log("Got:", v);
	    emitOut(3 * v);
	  }
	`)
	g.Feed("js.In", 123)
	g.Run()
	// Output:
	// Howdy from Otto!
	// Got: 123
	// Lost float64: 369
}
