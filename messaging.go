package jeebus

import (
	"net"

	proto "github.com/huin/mqtt"
	"github.com/jeffallen/mqtt"
)

var (
	mqttServer  *mqtt.Server
	mqttClient  *mqtt.ClientConn
	services    = make(map[string]Service)
	mqttStarted bool
)

// StartMessaging starts the internal MQTT messaging server and client.
func StartMessaging() {
	if mqttStarted {
		return
	}
	mqttStarted = true

	listener, err := net.Listen("tcp", ":1883")
	Check(err)

	mqttServer = mqtt.NewServer(listener)
	mqttServer.Start()
	// <-mqttServer.Done

	sock, err := net.Dial("tcp", ":1883")
	Check(err)

	mqttClient = mqtt.NewClientConn(sock)
	err = mqttClient.Connect("", "")
	Check(err)

	// subscribe to patterns which have already been setup (init code, etc)
	for pattern, _ := range services {
		addSubscription(pattern)
	}

	go func() {
		for m := range mqttClient.Incoming {
			Dispatch(m.TopicName, []byte(m.Payload.(proto.BytesPayload)))
		}
	}()
}

func Dispatch(topic string, payload []byte) {
	for pattern, handler := range services {
		if MatchTopic(pattern, topic) {
			handler.Handle(topic, payload)
		}
	}
}

func MatchTopic(pattern, topic string) bool {
	var skip = 0
	for i, p := range pattern {
		var n = i + skip
		switch {
		case p == '#':
			return true
		case n >= len(topic):
			return pattern[i:] == "/#"
		case byte(p) == topic[n]:
			continue
		case p == '+':
			for n < len(topic) && topic[n] != '/' {
				n++
			}
			skip = n - i - 1
		default:
			return false
		}
	}
	return len(pattern)+skip == len(topic)
}

type Service interface {
	Handle(topic string, payload []byte)
}

func Publish(topic string, payload interface{}) {
	mqttClient.Publish(&proto.Publish{
		Header:    proto.Header{Retain: topic[0] == '/'},
		TopicName: topic,
		Payload:   proto.BytesPayload(ToJson(payload)),
	})
}

func IsListeningTo(pattern string) bool {
	for k, _ := range services {
		if MatchTopic(pattern, k) {
			return true
		}
	}
	return false
}

func IsRegistered(pattern string) bool {
	_, ok := services[pattern]
	return ok
}

func Register(pattern string, service Service) {
	services[pattern] = service
	if mqttClient != nil {
		addSubscription(pattern)
	}
}

func Unregister(pattern string) {
	delete(services, pattern)
	// TODO: unsubscribe!
}

func addSubscription(pattern string) {
	mqttClient.Subscribe([]proto.TopicQos{{
		Topic: pattern,
		Qos:   proto.QosAtMostOnce,
	}})
}
