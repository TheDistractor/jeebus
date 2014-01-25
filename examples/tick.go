// Example of a ticking service for JeeBus, triggering itself
package main

import (
	"log"
	"time"

	"github.com/jcw/jeebus"
)

var (
	zzzBus jeebus.Client
)

type TickService int

func (s *TickService) Handle(tail string, value interface{}) {
	log.Printf("ZZZ '%s', value %#v (%T)", tail, value, value)
	*s = TickService(value.(float64))
}

func main() {
	zzzBus.Connect("zzz")
	zzzBus.Register("tick/foo", new(TickService))

	go func() {
		for {
			zzzBus.Publish("zzz/tick", 1.1)         // accepted, via broadcast
			zzzBus.Publish("zzz/tick/foo", 2.2)     // accepted, exact match
			zzzBus.Publish("zzz/tick/foo/bar", 3.3) // ignored
			zzzBus.Publish("zzz/tick/bar", 4.4)     // ignored
			zzzBus.Publish("zzz/bar", 5.5)          // ignored

			time.Sleep(3 * time.Second)
		}
	}()

	for {
		zzzBus.Publish("zzz/clock", time.Now()) // not picked up, just demo
		time.Sleep(time.Second)
	}
}
