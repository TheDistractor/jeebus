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
	T string          // topic
	P json.RawMessage // payload
	R bool            // retain
}

var (
	pubChan chan *Message
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

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
	Handle(subtopic string, value json.RawMessage)
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
			srvName := m.T[skip:]

			// look for an exact service match
			if service, ok := c.Services[srvName]; ok {
				service.Handle("", m.P)
			} else {
				// look for prefixes and wildcards
				srvPrefix := srvName + "/"
				for k, v := range c.Services {
					// if strings.HasPrefix(k, srvPrefix) {
					// 	v.Handle(k[len(srvPrefix):], m.P)
					// }
					n := len(k) - 1
					switch {
					//  pub "foo/bar" => sub "foo/bar/bleep"
					case strings.HasPrefix(k, srvPrefix):
						v.Handle(k[len(srvPrefix):], m.P)
					//  pub "foo/bar/bleep" => sub "foo/bar/#"
					case n >= 0 && k[n] == '#' && k[:n] == srvPrefix[:n]:
						v.Handle(srvName[n:], m.P)
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
		pubChan <- &Message{topic, v, retain}
	case json.RawMessage:
		pubChan <- &Message{topic, v, retain}
	default:
		data, err := json.Marshal(value)
		check(err)
		// log.Println("PUB", topic, string(data))
		pubChan <- &Message{topic, data, retain}
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
