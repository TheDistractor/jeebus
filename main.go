// This application exercises the "flow" package via a JSON config file.
// Use the "-i" flag for a list of built-in (i.e. pre-registered) gadgets.
package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/jcw/flow"
	"github.com/jcw/jeebus/gadgets"
)

var (
	verbose   = flag.Bool("i", false, "show info about version and registry")
	setupFile = flag.String("s", "setup.json", "name of the circuit setup file")
)

func init() {
	jeebus.Help["main"] = `Run the default circuit defined in the setup file.`
}

func main() {
	flag.Parse()

	err := flow.AddToRegistry(*setupFile)
	if err != nil && !*verbose {
		glog.Fatal(err)
	}

	if *verbose {
		println("JeeBus", jeebus.Version, "+ Flow", flow.Version, "\n")
		flow.PrintRegistry()
		println("\nUse 'help' for a list of commands or '-h' for a list of options.")
		println("Documentation at http://godoc.org/github.com/jcw/jeebus")
	} else {
		defer glog.Flush()
		glog.Infof("JeeBus %s - starting, registry size %d, args: %q",
			jeebus.Version, len(flow.Registry), flag.Args())

		appMain := flag.Arg(0)
		if appMain == "" {
			appMain = "main"
		}
		if factory, ok := flow.Registry[appMain]; ok {
			factory().Run()
		} else {
			glog.Fatalln(appMain, "not found in:", *setupFile)
		}
		glog.Infof("JeeBus %s - normal exit", jeebus.Version)
	}
}
