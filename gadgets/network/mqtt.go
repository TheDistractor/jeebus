// Interface to MQTT as client and as server.
package network

import (
	"encoding/json"
	"net"

	"github.com/golang/glog"
	proto "github.com/huin/mqtt"
	"github.com/jcw/flow"
	"github.com/jeffallen/mqtt"
)

func init() {
	flow.Registry["MQTTSub"] = func() flow.Circuitry { return &MQTTSub{} }
	flow.Registry["MQTTPub"] = func() flow.Circuitry { return &MQTTPub{} }
	flow.Registry["MQTTServer"] = func() flow.Circuitry { return &MQTTServer{} }
}

// MQTTSub can subscribe to an MQTT topic. Registers as "MQTTSub".
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

	topic := (<-w.Topic).(string)
	client.Subscribe([]proto.TopicQos{{
		Topic: topic,
		Qos:   proto.QosAtMostOnce,
	}})

	for m := range client.Incoming {
		payload := []byte(m.Payload.(proto.BytesPayload))
		var any interface{}
		err = json.Unmarshal(payload, &any)
		flow.Check(err)
		w.Out.Send(flow.Tag{m.TopicName, any})
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

	for m := range w.In {
		msg := m.(flow.Tag)
		glog.Infoln("MQTT publish", msg.Tag, msg.Msg)
		data, ok := msg.Msg.([]byte)
		if !ok {
			data, err = json.Marshal(msg.Msg)
			flow.Check(err)
		}
		retain := len(msg.Tag) > 0 && msg.Tag[0] == '/'
		client.Publish(&proto.Publish{
			Header:    proto.Header{Retain: retain},
			TopicName: msg.Tag,
			Payload:   proto.BytesPayload(data),
		})
	}
}

// MQTTServer is an embedded MQTT server. Registers as "MQTTServer".
type MQTTServer struct {
	flow.Gadget
	Port flow.Input
	PortOut flow.Output
}

// Start the MQTT server.
func (w *MQTTServer) Run() {
	port := (<-w.Port).(string)
	listener, err := net.Listen("tcp", port)
	flow.Check(err)
	glog.Infoln("MQTT server started, port", port)
	server := mqtt.NewServer(listener)
	server.Start()
	w.PortOut.Send(port)
	<-server.Done
	glog.Infoln("MQTT server done, port", port)
}
