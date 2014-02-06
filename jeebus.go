// JeeBus is a messaging and data storage infrastructure for low-end hardware.
package jeebus

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/url"
	"time"

	proto "github.com/huin/mqtt"
	"github.com/jeffallen/mqtt"
)

// Client represents a group of MQTT topics used as services.
type Client struct {
	Mqtt     *mqtt.ClientConn   // connection to the MQTT server
	Services map[string]Service // map subscriptions to handlers
	Done     chan bool          // unblocks when the server connection is lost
	ID       string             // uniquely identifies this endpoint
	cbs      *CallbackService   // takes care of RPC's and their replies
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
func NewClient(murl *url.URL) *Client {
	var err error
	var session net.Conn

	if murl == nil {
		murl, err = url.Parse("tcp://127.0.0.1:1883")
		check(err)
	}
	switch murl.Scheme {
	case "tcp":
		session, err = net.Dial("tcp", murl.Host)
	case "ssl":
		session, err = tls.Dial("tcp", murl.Host, nil)
	default:
		log.Fatalln("unknown scheme:", murl)
	}
	check(err)

	mc := mqtt.NewClientConn(session)
	err = mc.Connect("", "")
	check(err)

	myAddr := "ip-" + session.LocalAddr().String()
	cbs := &CallbackService{nil, make(map[int]chan *Message), 0}
	c := &Client{mc, make(map[string]Service), make(chan bool), myAddr, cbs}
	cbs.client = c

	// subscribe to receive all callback messages meant for us
	c.Register("cb/"+myAddr, cbs)

	go func() {
		for m := range mc.Incoming {
			c.Dispatch(m.TopicName, []byte(m.Payload.(proto.BytesPayload)))
		}
		log.Println("server connection lost")
		c.Done <- true
	}()

	log.Println("client connected")

	// go func() {
	// 	time.Sleep(time.Second)
	// 	v, e := c.Call("echo", 1, 4.5, "blah")
	// 	log.Println("RPC reply:", v, e)
	// 	check(e)
	// }()

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

func (c *Client) Call(name string, args ...interface{}) (interface{}, error) {
	return c.cbs.SendRPC(name, args)
}

type CallbackService struct {
	client  *Client
	pending map[int]chan *Message
	seqNum  int
}

func (s *CallbackService) SendRPC(name string, args []interface{}) (
	result interface{}, err error) {

	s.seqNum++
	request := []interface{}{s.seqNum, name}
	log.Printf("RPC #%d %s %v", s.seqNum, name, args)
	s.client.Publish("sv/rpc/"+s.client.ID, append(request, args...))

	c := make(chan *Message)
	s.pending[s.seqNum] = c
	defer delete(s.pending, s.seqNum)

	select {
	case m := <-c:
		var any []interface{}
		err := json.Unmarshal(m.P, &any)
		check(err)
		result = any[1]
		if any[2] != nil {
			err = errors.New(any[2].(string))
		}
	case <-time.After(10 * time.Second):
		log.Println("RPC timeout:", name)
		err = errors.New("timeout on RPC call: " + name)
	}
	return
}

func (s *CallbackService) Handle(m *Message) {
	var any []json.RawMessage
	err := json.Unmarshal(m.P, &any)
	check(err)
	var seq float64
	err = json.Unmarshal(any[0], &seq)
	check(err)
	if c, ok := s.pending[int(seq)]; ok {
		c <- m
	} else {
		log.Println("spurious reply ignored:", seq, any)
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
