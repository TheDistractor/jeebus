// JeeBus is a messaging and data storage infrastructure for low-end hardware.
package jeebus

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	proto "github.com/huin/mqtt"
	"github.com/jeffallen/mqtt"
)

type Message struct {
	T   string                     // topic
	P   json.RawMessage            // payload
	R   bool                       // retain
	obj map[string]json.RawMessage // decoded payload object fields
}

func (m *Message) unpack(key string) json.RawMessage {
	if len(m.obj) == 0 && len(m.P) > 0 {
		err := json.Unmarshal(m.P, &m.obj)
		check(err)
	}
	return m.obj[key]
}

func (m *Message) Get(key string) (v string) {
	json.Unmarshal(m.unpack(key), &v)
	return
}

func (m *Message) GetInt(key string) int {
	var f float64
	json.Unmarshal(m.unpack(key), &f)
	return int(f)
}

var (
	pubChan chan *Message
)

// Client represents a group of MQTT topics used as services.
type Client struct {
	Prefix   string
	Sub      chan *Message
	Services map[string]Service
}

func (c *Client) String() string {
	return fmt.Sprintf("«Cl:%s,%d»", c.Prefix, len(c.Services))
}

// Service represents the registration for a specific subtopic
type Service interface {
	// Handle gets called on the topic(s) it has been registered for.
	Handle(m *Message)
}

// Connect sets up a new MQTT connection for a specified client prefix.
func (c *Client) Connect(prefix string) {
	c.Prefix = prefix
	c.Sub = ConnectToServer(prefix + "/#")
	c.Services = make(map[string]Service)

	Publish("@/connect", prefix)
	log.Println("client connected:", prefix)

	go func() {
		// can't do this, since the connection has already been lost
		// defer Publish("@/disconnect", prefix)

		skip := len(prefix) + 1
		for m := range c.Sub {
			subTopic := m.T[skip:]
			message := &Message{T: subTopic, P: m.P}

			// TODO full MQTT wildcard match logic, i.e. also +'s
			// look for an exact service match
			if service, ok := c.Services[subTopic]; ok {
				service.Handle(message)
			} else {
				// look for prefixes and wildcards
				subPrefix := subTopic + "/"
				for k, v := range c.Services {
					n := len(k) - 1
					switch {
					//  pub "foo/bar" matches sub "foo/bar/bleep"
					case strings.HasPrefix(k, subPrefix):
						v.Handle(message)
					//  pub "foo/bar/bleep" matches sub "foo/bar/#"
					case n >= 0 && k[n] == '#' && k[:n] == subPrefix[:n]:
						v.Handle(message)
					}
				}
			}
		}

		log.Println("client disconnected:", prefix)
	}()
}

// Register a new service for a client with a specific prefix (can end in "#")
func (c *Client) Register(name string, service Service) {
	c.Services[name] = service
	Publish("@/register"+"/"+c.Prefix, name)
}

// Unregister a previously defined service.
func (c *Client) Unregister(name string) {
	Publish("@/unregister"+"/"+c.Prefix, name)
	delete(c.Services, name)
}

// Publish an arbitrary value to an arbitrary topic.
func Publish(topic string, value interface{}) {
	retain := topic[0] == '/'
	switch v := value.(type) {
	case []byte:
		pubChan <- &Message{T: topic, P: v, R: retain}
	case json.RawMessage:
		pubChan <- &Message{T: topic, P: v, R: retain}
	default:
		data, err := json.Marshal(value)
		check(err)
		// log.Println("PUB", topic, string(data))
		pubChan <- &Message{T: topic, P: data, R: retain}
	}
}

// ConnectToServer sets up an MQTT client and subscribes to the given topic(s).
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
					Header:    proto.Header{Retain: msg.R},
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
				R: m.Header.Retain,
			}
		}
		log.Println("server connection lost")
		close(sub)
	}()

	return sub
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
