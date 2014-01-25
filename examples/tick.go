// Example of a ticking service for JeeBus, triggering itself
package main

import (
	"log"
	"time"

	"github.com/jcw/jeebus"
)

type TickClient struct {
	jeebus.Client
}

type TickService int

func (s *TickService) Handle(c *jeebus.Client, tail string, value interface{}) {
	log.Printf("Svc zz: '%s', value %#v (%T)", tail, value, value)
	*s = TickService(value.(float64))
}

func main() {
	var client TickClient
	client.Connect("zz")
	client.Register("tick/foo", new(TickService))

	go func() {
		for {
			client.Publish("zz/tick", 1.1)         // accepted, via broadcast
			client.Publish("zz/tick/foo", 2.2)     // accepted, exact match
			client.Publish("zz/tick/foo/bar", 3.3) // ignored
			client.Publish("zz/tick/bleep", 4.4)   // ignored
			client.Publish("zz/bleep", 5.5)        // ignored

			time.Sleep(3 * time.Second)
		}
	}()

	for {
		client.Publish("zz/boom", time.Now().UnixNano()) // sent out with prefix
		time.Sleep(time.Second)
	}
}
