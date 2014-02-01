// JeeBus is a messaging and data storage infrastructure for low-end hardware.
package jeebus

import (
	"encoding/json"
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
			c.Dispatch(m.TopicName, m.Payload.(proto.BytesPayload))
		}
		log.Println("server connection lost")
		c.Done <- true
	}()

	// Publish("@/connect", prefix)
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
	c.Publish("@/register", topic)
}

// Unregister a previously defined service.
func (c *Client) Unregister(topic string) {
	c.Publish("@/unregister", topic)
	delete(c.Services, topic)
}

type Payload []byte

func (p *Payload) MarshalJSON() ([]byte, error) {
	log.Printf("Marshal %v", p)
	return *p, nil
}

func (p *Payload) UnmarshalJSON(v []byte) (error) {
	log.Printf("Unarshal %v", v)
	*p = v
	return nil
}

// NewPayload constructs a payload from just about any type of data.
func NewPayload(value interface{}) Payload {
	switch v := value.(type) {
	case []byte:
		return v
	case Payload:
		return []byte(v)
	case json.RawMessage:
		return []byte(v)
	default:
		data, err := json.Marshal(value)
		check(err)
		return data
	}
}

// Publish an arbitrary value to an arbitrary topic.
func (c *Client) Publish(topic string, value interface{}) {
	c.Mqtt.Publish(&proto.Publish{
		Header:    proto.Header{Retain: topic[0] == '/'},
		TopicName: topic,
		Payload:   proto.BytesPayload(NewPayload(value)),
	})
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
