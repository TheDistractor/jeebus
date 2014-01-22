package main

import (
	"log"
	"os"

	".." // the fully-specified path is "github.com/jcw/jeebus"
)

func main() {
	switch jeebus.SubCommand("jb") {
  // no extra sub-commands defined
	default:
		log.Fatal("unknown sub-command: jb ", os.Args[1], " ...")
	}
}
