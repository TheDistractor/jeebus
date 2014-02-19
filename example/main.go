package main

import (
	"log"

	"github.com/jcw/jeebus"
)

func init() {
	log.SetFlags(log.Ltime)
}

func main() {
	log.Println("JeeBus example", jeebus.Version)
	jeebus.Run()
}
