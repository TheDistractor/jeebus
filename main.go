// This application exercises the "flow" package via a JSON config file.
// Use the "-i" flag for a list of built-in (i.e. pre-registered) gadgets.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/jcw/flow"
	"github.com/jcw/jeebus/gadgets"
)

var (
	verbose   = flag.Bool("i", false, "show info about version and registry")
	setupFile = flag.String("s", "setup.json", "name of the circuit setup file")
)

func main() {
	flag.Parse()

	// special info if caller is node.js, to pass the PID of this process
	// yes, this is *writing* to stdin (which is used as IPC mechanism!)
	if flag.NArg() == 0 {
		// hack alert: only do this in the default case, i.e. without args
		// TODO: need a way to detect whether node.js launched this app!
		// perhaps check whether stdin is a pipe? is this portable?
		// or check that the raw stdin and stdout fd's are different
		os.Stdin.Write([]byte(fmt.Sprintf("%d\n", os.Getpid())))
	}
	// see the websocket code for how input from stdin is picked up

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
		glog.Infof("JeeBus %s - starting, registry size %d, args: %v",
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
		glog.Infof("JeeBus %s -, normal exit", jeebus.Version)
	}
}
