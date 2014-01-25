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
	mqttClient *mqtt.ClientConn // TODO get rid of this
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Client represents a group of MQTT topics used as services.
type Client struct {
	Prefix   string
	Pub, Sub chan *Message
	Services map[string]Service
}

func (c *Client) String() string {
    return fmt.Sprintf("«Cl:%s,%d»", c.Prefix, len(c.Services))
}

// Service represents the registration for a specific subtopic
type Service interface {
    // Handle gets called on the topic(s) it has been registered for
    Handle(c *Client, subtopic string, value interface{})
}

// Connect sets up a new MQTT connection for a specified client prefix.
func (c *Client) Connect(prefix string) {
    c.Prefix = prefix
	c.Pub, c.Sub = ConnectToServer(prefix + "/#")
    c.Services = make(map[string]Service)

    // client := &Client{prefix, pub, sub, make(map[string]Service)}
	c.Publish("@/connect", prefix)
	log.Println("client connected:", prefix)

	go func() {
		// can't do this, since the connection has already been lost
		// defer client.Publish("@/disconnect", prefix)

		skip := len(prefix) + 1
		for m := range c.Sub {
			srvName := m.T[skip:]
			var value interface{}
			err := json.Unmarshal(m.P, &value)
			check(err)

			// first look for the special "#" wildcard
			check(err)
			if service, ok := c.Services["#"]; ok {
				service.Handle(c, srvName, value)
			}

			// then look for an exact service match
			if service, ok := c.Services[srvName]; ok {
				service.Handle(c, "", value)
			}

			// finally look for all services which are a prefix of this topic
			srvPrefix := srvName + "/"
			for k, v := range c.Services {
				if strings.HasPrefix(k, srvPrefix) {
					v.Handle(c, k[len(srvPrefix):], value)
				}
			}
		}

		log.Println("client disconnected:", prefix)
	}()
}

// Register a new service for a client, using a more specific prefix.
func (c *Client) Register(name string, service Service) {
	c.Services[name] = service
	c.Publish("@/register"+"/"+c.Prefix, name)
}

// Unregister a previously defined service.
func (c *Client) Unregister(name string) {
	c.Publish("@/unregister"+"/"+c.Prefix, name)
	delete(c.Services, name)
}

// Publish an arbitrary value to an arbitrary topic.
func (c *Client) Publish(topic string, value interface{}) {
	switch value := value.(type) {
	case []byte:
		c.Pub <- &Message{T: topic, P: value}
	default:
		data, err := json.Marshal(value)
		check(err)
		c.Pub <- &Message{T: topic, P: data}
	}
}

func ConnectToServer(topic string) (pub, sub chan *Message) {
    session, err := net.Dial("tcp", "localhost:1883")

	mqttClient = mqtt.NewClientConn(session)
	err = mqttClient.Connect("", "")
	check(err)

	mqttClient.Subscribe([]proto.TopicQos{
		{Topic: topic, Qos: proto.QosAtMostOnce},
	})

    // TODO don't need one for each client, just one per connection
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
        // close(sub)
	}()

	return
}
