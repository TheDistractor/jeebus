package main

import (
	"log"
	"os"

	"github.com/jcw/jeebus"
)

func init() {
	log.SetFlags(log.Ltime)
}

func main() {
	log.Println("JeeBus example", jeebus.Version, "pid", os.Getpid())
	jeebus.Run()
}
