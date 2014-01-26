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

// Message represent a payload over MQTT for a specified topic.
type Message struct {
	T   string                     // topic
	P   json.RawMessage            // payload
	R   bool                       // retain
	obj map[string]json.RawMessage // decoded payload object fields
}

// String returns a short string representation of a Message.
func (m *Message) String() string {
	// display the retain flag only if set
	retain := ""
	if m.R {
		retain = ",R"
	}
	// insert an ellipsis if the payload data is too long
	// note that all numbers, booleans, and nulls will pass through as is
	msg := string(m.P)
	if len(msg) > 20 {
		msg = msg[:18] + "…"
		switch msg[0] {
		case '{':
			msg += "}"
		case '[':
			msg += "]"
		default:
			msg += msg[:1] // only double quotes, really
		}
	}
	// replace the most common non-printable characters by a dot
	f := func(r rune) rune {
		if r < ' ' {
			r = '.'
		}
		return r
	}
	return fmt.Sprintf("«M:%s,%s%s»", m.T, strings.Map(f, msg), retain)
}

// unpack the JSON payload into a map, this fails if payload is not an object.
func (m *Message) unpack(key string) json.RawMessage {
	if len(m.obj) == 0 && len(m.P) > 0 {
		err := json.Unmarshal(m.P, &m.obj)
		check(err)
	}
	return m.obj[key]
}

// Get extracts a given object attribute as string, or "" if absent.
func (m *Message) Get(key string) (v string) {
	json.Unmarshal(m.unpack(key), &v)
	return
}

// GetBool extracts a given object attribute as bool, or false if absent.
func (m *Message) GetBool(key string) (v bool) {
	json.Unmarshal(m.unpack(key), &v)
	return
}

// GetInt extracts a given object attribute as int, or 0 if absent.
func (m *Message) GetInt(key string) int {
	return int(m.GetFloat64(key))
}

// GetInt64 extracts a given object attribute as 64-bit int, or 0 if absent.
func (m *Message) GetInt64(key string) int64 {
	return int64(m.GetFloat64(key))
}

// GetFloat64 extracts a given object attribute as float, or 0 if absent.
func (m *Message) GetFloat64(key string) (v float64) {
	json.Unmarshal(m.unpack(key), &v)
	return
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

// String returns a short string representation of a Client.
func (c *Client) String() string {
	return fmt.Sprintf("«C:%s,%d»", c.Prefix, len(c.Services))
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
