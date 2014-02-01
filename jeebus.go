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
	Done	 chan bool
}

func (c *Client) Dispatch(m *Message) {
	// TODO full MQTT wildcard match logic, i.e. also +'s
	topic := []byte(m.T)

	for k, v := range c.Services {
		if k == m.T {
			v.Handle(m)
		} else {
			for i, b := range []byte(k) {
				if b == '#' || i == len(topic) && b == '/' {
					v.Handle(m)
					break
				}
			}
		}
	}

/*
	// look for an exact service match
	if service, ok := c.Services[subTopic]; ok {
		service.Handle(message)
	} else {
		// look for prefixes and wildcards
		subPrefix := subTopic + "/"
		for k, v := range c.Services {
			n := len(k) - 1
			// log.Printf("SLICE n %d k %s sp %s", n, k, subPrefix)
			switch {
			//  pub "foo/bar" matches sub "foo/bar/bleep"
			case strings.HasPrefix(k, subPrefix):
				v.Handle(message)
			//  pub "foo/bar/bleep" matches sub "foo/bar/#"
			case n >= 0 && n <= len(subPrefix) &&
				k[n] == '#' && k[:n] == subPrefix[:n]:
				v.Handle(message)
			}
		}
	}
*/
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
			payload := json.RawMessage(m.Payload.(proto.BytesPayload))
			c.Dispatch(&Message{T: m.TopicName, P: payload})
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

// Publish an arbitrary value to an arbitrary topic.
func (c *Client) Publish(topic string, value interface{}) {
	var m *Message
	switch v := value.(type) {
	case []byte:
		m = &Message{T: topic, P: v}
	case json.RawMessage:
		m = &Message{T: topic, P: v}
	default:
		data, err := json.Marshal(value)
		check(err)
		// log.Println("PUB", topic, string(data))
		m = &Message{T: topic, P: data}
	}
	c.Mqtt.Publish(&proto.Publish{
		Header:    proto.Header{Retain: m.T[0] == '/'},
		TopicName: m.T,
		Payload:   proto.BytesPayload(m.P),
	})
}

// ConnectToServer sets up an MQTT client and subscribes to the given topic(s).
/*
func ConnectToServer(topic string) chan *Message {
	session, err := net.Dial("tcp", "localhost:1883")

	mqttClient := mqtt.NewClientConn(session)
	err = mqttClient.Connect("", "")
	check(err)

	mqttClient.Subscribe([]proto.TopicQos{
		{Topic: topic, Qos: proto.QosAtMostOnce},
	})

	// set up a channel to publish through, but only once
	if pubChan == nil {
		pubChan = make(chan *Message)
		go func() {
			for msg := range pubChan {
				mqttClient.Publish(&proto.Publish{
					Header:    proto.Header{Retain: msg.T[0] == '/'},
					TopicName: msg.T,
					Payload:   proto.BytesPayload(msg.P),
				})
			}
		}()
	}

	sub := make(chan *Message)
	go func() {
		for m := range mqttClient.Incoming {
			sub <- &Message{
				T: m.TopicName,
				P: json.RawMessage(m.Payload.(proto.BytesPayload)),
				// R: m.Header.Retain,
			}
		}
		log.Println("server connection lost")
		close(sub)
	}()

	return sub
}
*/

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
