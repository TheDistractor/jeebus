// JeeBus is a messaging and data storage infrastructure for low-end hardware.
package jeebus

import (
	"encoding/json"
	"log"
	"net"

	proto "github.com/huin/mqtt"
	"github.com/jeffallen/mqtt"
)

type Message struct {
	T string          // topic
	P json.RawMessage // payload
	R bool            // retain
}

var (
	mqttClient *mqtt.ClientConn // TODO get rid of this
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Client struct {
	Tag      string
	Sub, Pub chan *Message
	Handlers map[string]ClientService
}

type ClientService interface {
}

func NewClient(tag string) *Client {
	sub, pub := ConnectToServer(":" + tag + "/#")
	return &Client{
		Tag:      tag,
		Sub:      sub,
		Pub:      pub,
		Handlers: make(map[string]ClientService),
	}
}

// func (c *Client) AddService(name string) {
// 	c.handlers[name] = 1
// }
//
// func (c *Client) RemoveService(name string) {
// 	delete(handlers, name)
// }

func (c *Client) Publish(key string, value interface{}) {
	topic := c.Tag + "/" + key
	switch value := value.(type) {
	case []byte:
		c.Pub <- &Message{T: topic, P: value}
	default:
		data, err := json.Marshal(value)
		check(err)
		c.Pub <- &Message{T: topic, P: data}
	}
}

func ConnectToServer(topic string) (sub, pub chan *Message) {
	conn, err := net.Dial("tcp", "localhost:1883")
	check(err)

	mqttClient = mqtt.NewClientConn(conn)
	err = mqttClient.Connect("", "")
	check(err)

	mqttClient.Subscribe([]proto.TopicQos{
		{Topic: topic, Qos: proto.QosAtMostOnce},
	})

	// set up a channel to publish through
	pub = make(chan *Message)
	go func() {
		for msg := range pub {
			mqttClient.Publish(&proto.Publish{
				Header:    proto.Header{Retain: msg.R},
				TopicName: msg.T,
				Payload:   proto.BytesPayload(msg.P),
			})
		}
	}()

	sub = make(chan *Message)
	go func() {
		for m := range mqttClient.Incoming {
			sub <- &Message{
				T: m.TopicName,
				P: json.RawMessage(m.Payload.(proto.BytesPayload)),
				R: m.Header.Retain,
			}
		}
		log.Println("server connection lost")
		close(sub)
	}()

	return
}
