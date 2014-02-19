package jeebus

import (
	"net"
	"strings"

	proto "github.com/huin/mqtt"
	"github.com/jeffallen/mqtt"
)

var (
	mqttServer *mqtt.Server
	mqttClient *mqtt.ClientConn
	services   = make(map[string]Service)
)

// StartMessaging starts the internal MQTT messaging server and client.
func StartMessaging(listener net.Listener) error {
	var err error

	if listener == nil {
		if listener, err = net.Listen("tcp", ":1883"); err != nil {
			return err
		}
	}

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
			topic := m.TopicName
			data := []byte(m.Payload.(proto.BytesPayload))
			var body interface{} = data

			// FIXME special-cased to treat "io/#" topics as binary data
			if !strings.HasPrefix(topic, "io/") {
				body = FromJson(data)
			}

			for pattern, handler := range services {
				if MatchTopic(pattern, topic) {
					handler.Handle(topic, body)
				}
			}
		}
	}()

	return nil
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
	Handle(topic string, payload interface{})
}

func Publish(topic string, payload interface{}) {
	mqttClient.Publish(&proto.Publish{
		Header:    proto.Header{Retain: topic[0] == '/'},
		TopicName: topic,
		Payload:   proto.BytesPayload(ToJson(payload)),
	})
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
