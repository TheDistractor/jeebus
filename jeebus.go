// JeeBus is a messaging and data storage infrastructure for low-end hardware.
package jeebus

import (
	"log"
	"net"

	proto "github.com/huin/mqtt"
	"github.com/jeffallen/mqtt"
)

// Client represents a group of MQTT topics used as services.
type Client struct {
	Mqtt     *mqtt.ClientConn
	Services map[string]Service
	Done     chan bool
}

// Dispatch a payload to the appropriate registered services for that topic.
func (c *Client) Dispatch(topic string, payload interface{}) {
	// TODO full MQTT wildcard match logic, i.e. also +'s
	m := &Message{T: topic, P: NewPayload(payload)}
	t := []byte(topic)

	for k, v := range c.Services {
		if k == topic {
			v.Handle(m)
		} else {
			for i, b := range []byte(k) {
				if b == '#' || i == len(t) && b == '/' {
					v.Handle(m)
				} else if b == t[i] {
					continue
				}
				break
			}
		}
	}
}

// Service represents the registration for a specific subtopic
type Service interface {
	// Handle gets called on the topic(s) it has been registered for.
	Handle(m *Message)
}

// NewClient sets up a new MQTT connection plus registration mechanism
func NewClient() *Client {
	session, err := net.Dial("tcp", "localhost:1883")

	mc := mqtt.NewClientConn(session)
	err = mc.Connect("", "")
	check(err)

	c := &Client{mc, make(map[string]Service), make(chan bool)}

	go func() {
		for m := range mc.Incoming {
			c.Dispatch(m.TopicName, []byte(m.Payload.(proto.BytesPayload)))
		}
		log.Println("server connection lost")
		c.Done <- true
	}()

	log.Println("client connected")

	return c
}

// Register a new service for a client with a specific prefix (can end in "#")
func (c *Client) Register(topic string, service Service) {
	if _, ok := c.Services[topic]; ok {
		log.Fatal("canno register service twice:", topic)
	}
	c.Services[topic] = service
	c.Mqtt.Subscribe([]proto.TopicQos{
		{Topic: topic, Qos: proto.QosAtMostOnce},
	})
}

// Unregister a previously defined service.
func (c *Client) Unregister(topic string) {
	// TODO unsubscribe...
	delete(c.Services, topic)
}

// Publish an arbitrary payload to the specified topic.
func (c *Client) Publish(topic string, payload interface{}) {
	c.Mqtt.Publish(&proto.Publish{
		Header:    proto.Header{Retain: topic[0] == '/'},
		TopicName: topic,
		Payload:   proto.BytesPayload(NewPayload(payload)),
	})
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
