package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jcw/jeebus"
	"github.com/jeffallen/mqtt"
)

func startMqttServer() chan *jeebus.Message {
	port, err := net.Listen("tcp", ":1883")
	check(err)
	svr := mqtt.NewServer(port)
	svr.Start()
	// <-svr.Done

	// TODO everything below can be dropped once all routing works properly

	pub, sub := jeebus.ConnectToServer("#")

	go func() {
		for m := range sub {
			topic := m.T
			message := []byte(m.P)

			// FIXME can't work: retain flag is not published to subscribers!
			//	solving this will require a modified mqtt package
			// if m.Header.Retain {
			// 	store("mqtt/"+topic, message, nil)
			// }

			switch topic[:3] {

			// st/key... -> current state, stored as key and with timestamp
			case "st/":
				key := topic[3:]
				store(key, message)
				millis := time.Now().UnixNano() / 1000000
				store(fmt.Sprintf("hist/%s/%d", key, millis), message)

			// db/... -> database requests, value is reply topic
			case "db/":
				if strings.HasPrefix(topic, "db/get/") {
					pub <- &jeebus.Message{
						T: string(message),
						P: fetch(topic[7:]),
					}
				}
			}
		}
	}()

	return pub
}
