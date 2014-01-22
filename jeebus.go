package jeebus

import (
	"encoding/json"
	"log"
	"net"

	proto "github.com/huin/mqtt"
	"github.com/jeffallen/mqtt"
)

type Message struct {
	T string      // topic
	P interface{} // payload
	R bool        // retain
}

var (
	mqttClient *mqtt.ClientConn // TODO get rid of this
	PubChan    chan *Message
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// TODO get rid of this, use PubChan and add support for raw sending
func Publish(key string, value []byte) {
	// log.Printf("P %s => %s", key, value)
	mqttClient.Publish(&proto.Publish{
		// Header:    proto.Header{Retain: true},
		TopicName: key,
		Payload:   proto.BytesPayload(value),
	})
}

func ListenToServer(topic string) chan Message {
	conn, err := net.Dial("tcp", "localhost:1883")
	check(err)

	mqttClient = mqtt.NewClientConn(conn)
	err2 := mqttClient.Connect("", "")
	check(err2)

	mqttClient.Subscribe([]proto.TopicQos{
		{Topic: topic, Qos: proto.QosAtMostOnce},
	})

	// set up a channel to publish through
	PubChan = make(chan *Message)
	go func() {
		for msg := range PubChan {
			// log.Printf("C %s => %v", msg.T, msg.P)
			value, err := json.Marshal(msg.P)
			check(err)
			mqttClient.Publish(&proto.Publish{
				Header:    proto.Header{Retain: msg.R},
				TopicName: msg.T,
				Payload:   proto.BytesPayload(value),
			})
		}
	}()

	listenChan := make(chan Message)
	go func() {
		for m := range mqttClient.Incoming {
			listenChan <- Message{
				T: m.TopicName,
				P: []byte(m.Payload.(proto.BytesPayload)),
				R: m.Header.Retain,
			}
		}
		log.Println("server connection lost")
		close(listenChan)
	}()

	return listenChan
}
