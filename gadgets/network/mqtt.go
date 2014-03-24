// Interface to MQTT as client and as server.
package network

// glog levels:
//	1 = publish
//  2 = subscribe

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

func getInputOrConfig(vin flow.Input, vname string) string {
	// if a port is given, use it, else use the default from the configuration
	value := flow.Config[vname]
	if m := <-vin; m != nil {
		value = m.(string)
	}
	if value == "" {
		glog.Errorln("no value given for:", vname)
	}
	return value
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
	port := getInputOrConfig(w.Port, "MQTT_PORT")
	sock, err := net.Dial("tcp", port)
	flow.Check(err)
	client := mqtt.NewClientConn(sock)
	err = client.Connect("", "")
	flow.Check(err)

	for t := range w.Topic {
		topic := t.(string)
		glog.V(2).Infoln("mqtt-sub", topic)
		client.Subscribe([]proto.TopicQos{{
			Topic: topic,
			Qos:   proto.QosAtMostOnce,
		}})
	}

	for m := range client.Incoming {
		payload := []byte(m.Payload.(proto.BytesPayload))
		// try to decode as JSON, but leave as []byte if that fails
		var any interface{}
		if err = json.Unmarshal(payload, &any); err != nil {
			any = payload
		}
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
	port := getInputOrConfig(w.Port, "MQTT_PORT")
	sock, err := net.Dial("tcp", port)
	flow.Check(err)
	client := mqtt.NewClientConn(sock)
	err = client.Connect("", "")
	flow.Check(err)

	for m := range w.In {
		msg := m.(flow.Tag)
		glog.V(1).Infoln("mqtt-pub", msg.Tag, msg.Msg)
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
	Port    flow.Input
	PortOut flow.Output
}

// Start the MQTT server.
func (w *MQTTServer) Run() {
	port := getInputOrConfig(w.Port, "MQTT_PORT")
	listener, err := net.Listen("tcp", port)
	flow.Check(err)
	glog.Infoln("mqtt started, port", port)
	server := mqtt.NewServer(listener)
	server.Start()
	w.PortOut.Send(port)
	<-server.Done
	glog.Infoln("mqtt done, port", port)
}
