package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jcw/jeebus" // TODO remove dependency
	"github.com/jeffallen/mqtt"
)

var (
	mqttClient *mqtt.ClientConn // TODO get rid of this
)

func startMqttServer() {
	port, err := net.Listen("tcp", ":1883")
	check(err)
	svr := mqtt.NewServer(port)
	svr.Start()
	// <-svr.Done

	feed := jeebus.ListenToServer("#")

	jeebus.Publish("st/admin/started", []byte(time.Now().Format(time.RFC822Z)))

	go func() {
		for m := range feed {
			mqttDispatch(m)
		}
	}()
}

func mqttDispatch(m *jeebus.Message) {
	topic := m.T
	message := m.P.([]byte)

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
			value := fetch(topic[7:])
			jeebus.Publish(string(message), value)
		}

	// TODO hardcoded serial port to websocket pass-through for now
	case "if/":
		if strings.HasPrefix(topic, "if/serial/") {
			sendToAllWebSockets(message)
		}

	// TODO hardcoded websocket to serial port pass-through for now
	case "ws/":
		// accept arrays of arbitrary data types
		var any []interface{}
		err := json.Unmarshal(message, &any)
		check(err)
		// send as L<n><m> to the serial port
		cmd := fmt.Sprintf("L%.0f%.0f", any[0], any[1])
		jeebus.Publish("if/serial", []byte(cmd))
	}
}
