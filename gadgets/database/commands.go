package database

import (
	"flag"
	"fmt"
	
	"github.com/jcw/flow"
)

func init() {
	flow.Registry["dump"] = func() flow.Circuitry { return &dumpCmd{} }
}

type dumpCmd struct {
	flow.Gadget
}

func (g *dumpCmd) Run() {
	odb := openDatabase("./db")
	odb.iterateOverKeys(flag.Arg(1), flag.Arg(2), func(k string, v []byte) {
		fmt.Printf("%s = %s\n", k, v)
	})
}
