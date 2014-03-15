// This application exercises the "flow" package via a JSON config file.
// Use the "-i" flag for a list of built-in (i.e. pre-registered) gadgets.
package main

import (
	"flag"
	"time"

	"github.com/golang/glog"
	"github.com/jcw/flow"
	"github.com/jcw/jeebus/gadgets"
)

var (
	verbose   = flag.Bool("i", false, "show info about version and registry")
	wait      = flag.Bool("w", false, "wait forever, don't exit main")
	setupFile = flag.String("s", "setup.json", "circuitry setup file")
	appMain   = flag.String("r", "main", "which registered circuit to run")
)

func main() {
	flag.Parse()

	err := flow.AddToRegistry(*setupFile)
	if err != nil && !*verbose {
		glog.Fatal(err)
	}

	if *verbose {
		println("JeeBus", jeebus.Version, "+ Flow", flow.Version, "\n")
		flow.PrintRegistry()
		println("\nDocumentation at http://godoc.org/github.com/jcw/jeebus")
	} else {
		glog.Infof("JeeBus %s - starting, registry size %d",
			jeebus.Version, len(flow.Registry))
		if factory, ok := flow.Registry[*appMain]; ok {
			factory().Run()
			if *wait {
				println("waiting...")
				time.Sleep(1e6 * time.Hour)
			}
		} else {
			glog.Fatalln(*appMain, "not found in:", *setupFile)
		}
		glog.Infof("JeeBus %s -, normal exit", jeebus.Version)
	}
}
