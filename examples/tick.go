// Example of a ticking service for JeeBus, triggering itself.
package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/jcw/jeebus"
)

var (
	zzzBus jeebus.Client
)

type TickService int

func (s *TickService) Handle(m *jeebus.Message) {
	log.Printf("ZZZ subtopic '%s', payload %s", m.T, m.P)
	var num float64
	err := json.Unmarshal(m.P, &num)
	if err != nil {
		log.Fatal(err)
	}
	*s = TickService(num)
}

func main() {
	zzzBus.Connect("zzz")
	zzzBus.Register("tick/foo", new(TickService))

	go func() {
		for {
			jeebus.Publish("zzz/tick", 1.1)         // accepted, via broadcast
			jeebus.Publish("zzz/tick/foo", 2.2)     // accepted, exact match
			jeebus.Publish("zzz/tick/foo/bar", 3.3) // ignored
			jeebus.Publish("zzz/tick/bar", 4.4)     // ignored
			jeebus.Publish("zzz/bar", 5.5)          // ignored

			time.Sleep(3 * time.Second)
		}
	}()

	for {
		jeebus.Publish("zzz/clock", time.Now()) // not picked up, just demo
		time.Sleep(time.Second)
	}
}
