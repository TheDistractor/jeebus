// Example of a minimal application based on JeeBus.
package main

import (
	"github.com/jcw/flow"             // main dataflow package
	_ "github.com/jcw/jeebus/gadgets" // load additional gadgets
)

const demo = `{
	"gadgets": [
		{ "name": "c", "type": "Clock" },
		{ "name": "p", "type": "Printer" }
	],
	"wires": [{ "from": "c.Out", "to": "p.In" }],
	"feeds": [{ "data": "1s", "to": "c.Rate" }]
}`

func main() {
	c := flow.NewCircuit()
	c.LoadJSON([]byte(demo))
	c.Run()
}
