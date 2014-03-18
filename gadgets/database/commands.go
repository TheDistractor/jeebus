package database

import (
	"flag"
	"github.com/jcw/flow"
)

func init() {
	flow.Registry["dump"] = func() flow.Circuitry { return &dumpCmd{} }
}

type dumpCmd struct {
	flow.Gadget
}

func (g *dumpCmd) Run() {
	println(flag.Args())
}
