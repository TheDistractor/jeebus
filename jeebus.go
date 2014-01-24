// JeeBus is a messaging and data storage infrastructure for low-end hardware.
package jeebus

import (
	"encoding/json"
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

// Client represents a group of MQTT topics used as services
type Client struct {
	Prefix   string
	Pub, Sub chan *Message
	Services map[string]ClientService
}

// ClientService listen to specific topics, and can emit its own messages
type ClientService interface {
	Handle(topic string, value interface{})
}

// NewClient sets up a new MQTT connection for a specified client prefix
func NewClient(prefix string) *Client {
	pub, sub := ConnectToServer(":" + prefix + "/#")

	client := &Client{prefix, pub, sub, make(map[string]ClientService)}
	client.Publish(":@/connect", prefix)
	log.Println("client connected:", prefix)

	go func() {
		// can't do this, since the connection has already been lost
		// defer client.Publish(":@/disconnect", prefix)

		skip := len(prefix) + 2
		for m := range sub {
			// first look for an exact service match
			srvName := m.T[skip:]
			if srv, ok := client.Services[srvName]; ok {
				srv.Handle("", m.P)
			}
			// then look for all services which have this topic as prefix
			srvPrefix := srvName + "/"
			for k, v := range client.Services {
				if strings.HasPrefix(k, srvPrefix) {
					v.Handle(m.T[skip+1:], m.P)
				}
			}
		}

		log.Println("client disconnected:", prefix)
	}()

	return client
}

// Register a new service for a client, using a more specific prefix
func (c *Client) Register(name string, service ClientService) {
	c.Services[name] = service
	c.Publish(":@/register", []byte(name))
}

// Unregister a previously defined service
func (c *Client) Unregister(name string) {
	c.Publish(":@/unregister", []byte(name))
	delete(c.Services, name)
}

// Publish an arbitrary value to an arbitrary topic
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

// Emit (i.e. publish) an arbitrary value to a topic with this client's prefix
func (c *Client) Emit(key string, value interface{}) {
	c.Publish(c.Prefix+"/"+key, value)
}

func ConnectToServer(topic string) (pub, sub chan *Message) {
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
