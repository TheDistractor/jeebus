// Interface to MQTT as client and as server.
package network

import (
	"net"

	proto "github.com/huin/mqtt"
	"github.com/jcw/flow"
	"github.com/jeffallen/mqtt"
)

func init() {
	flow.Registry["MQTTSub"] = func() flow.Circuitry { return &MQTTSub{} }
	flow.Registry["MQTTPub"] = func() flow.Circuitry { return &MQTTPub{} }
	flow.Registry["MQTTServer"] = func() flow.Circuitry { return &MQTTServer{} }
}

// MQTTSub can subscribe to MQTT. Registers as "MQTTSub".
type MQTTSub struct {
	flow.Gadget
	Port  flow.Input
	Topic flow.Input
	Out   flow.Output
}

// Start listening and subscribing to MQTT.
func (w *MQTTSub) Run() {
	port := (<-w.Port).(string)
	sock, err := net.Dial("tcp", port)
	flow.Check(err)
	client := mqtt.NewClientConn(sock)
	err = client.Connect("", "")
	flow.Check(err)

	if topic, ok := <-w.Topic; ok {
		client.Subscribe([]proto.TopicQos{{
			Topic: topic.(string),
			Qos:   proto.QosAtMostOnce,
		}})
		for m := range client.Incoming {
			payload := []byte(m.Payload.(proto.BytesPayload))
			w.Out.Send(flow.Tag{m.TopicName, payload})
		}
	}
}

// MQTTPub can publish to MQTT. Registers as "MQTTPub".
type MQTTPub struct {
	flow.Gadget
	Port flow.Input
	In   flow.Input
}

// Start publishing to MQTT.
func (w *MQTTPub) Run() {
	port := (<-w.Port).(string)
	sock, err := net.Dial("tcp", port)
	flow.Check(err)
	client := mqtt.NewClientConn(sock)
	err = client.Connect("", "")
	flow.Check(err)

	if m, ok := <-w.In; ok {
		msg := m.([]string)
		client.Publish(&proto.Publish{
			Header:    proto.Header{Retain: msg[0][0] == '/'},
			TopicName: msg[0],
			Payload:   proto.BytesPayload(msg[1]),
		})
	}
}

// MQTTServer is an embedded MQTT server. Registers as "MQTTServer".
type MQTTServer struct {
	flow.Gadget
	Port flow.Input
}

// Start the MQTT server.
func (w *MQTTServer) Run() {
	port := (<-w.Port).(string)
	listener, err := net.Listen("tcp", port)
	flow.Check(err)
	server := mqtt.NewServer(listener)
	server.Start()
	<-server.Done
}
