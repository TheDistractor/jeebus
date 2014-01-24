package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jcw/jeebus"
	"github.com/jeffallen/mqtt"
)

var (
	mqttClient *mqtt.ClientConn // TODO get rid of this
)

func startMqttServer() chan *jeebus.Message {
	port, err := net.Listen("tcp", ":1883")
	check(err)
	svr := mqtt.NewServer(port)
	svr.Start()
	// <-svr.Done

	pub, sub := jeebus.ConnectToServer("#")

	pub <- &jeebus.Message{
		T: "st/admin/started",
		P: []byte(time.Now().Format(time.RFC822Z)),
	}

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

			// TODO hardcoded serial port to websocket pass-through for now
			case "if/":
				if strings.HasPrefix(topic, "if/serial/") {
					sendToAllWebSockets(message)
				}

			// TODO type sent is wrong and it shouldn't be limited to strings
			case "ws/":
				var any []string
				err := json.Unmarshal(message, &any)
				check(err)
				pub <- &jeebus.Message{
					T: any[0],
					P: []byte(any[1]),
				}
			}
		}
	}()

	return pub
}
