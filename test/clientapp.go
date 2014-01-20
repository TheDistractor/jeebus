package main

import (
	//"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	//"net/http"
	//"os"
	//"strconv"
	"strings"
	"time"

	proto "github.com/huin/mqtt"
	"github.com/jeffallen/mqtt"
)

type BusMessage struct {
	T string      // topic
	M interface{} // message
	R bool        // retain
}

var (
	mqttClient     *mqtt.ClientConn
	busPubChan     chan *BusMessage
)



func init() {
	busPubChan = make(chan *BusMessage)
}

func main() {
	log.Println("starting MQTT client")
	startmqttClient()
	log.Println("MQTT client is running")
	
	// publish simulated data
    ticker := time.NewTicker(time.Second * 5)
	topic := "data/time"
	for t := range ticker.C {
		fmt.Println("Tick at", t)
		busPubChan <- &BusMessage{T: topic, M: t}
	}
}

func startmqttClient() {
	go func() {
		//port, err := net.Listen("tcp", ":1883")
		//if err != nil {
		//	log.Fatal("listen: ", err)
		//}

		conn, _ := net.Dial("tcp", "localhost:1883")
		mqttClient = mqtt.NewClientConn(conn)
		// mqttClient.Dump = true
		mqttClient.Connect("", "")
		Publish("st/admin/started", []byte(time.Now().Format(time.RFC822Z)))

		mqttClient.Subscribe([]proto.TopicQos{
			{Topic: "#", Qos: proto.QosAtMostOnce},
		})

		for m := range mqttClient.Incoming {
			mqttHandleIncoming(m)
		}
	}()

	// set up a channel to publish through
	go func() {
		for msg := range busPubChan {
			log.Printf("C %s => %v", msg.T, msg.M)
			value, err := json.Marshal(msg.M)
			if err != nil {
				log.Fatal(msg, err)
			}
			mqttClient.Publish(&proto.Publish{
				Header:    proto.Header{Retain: msg.R},
				TopicName: msg.T,
				Payload:   proto.BytesPayload(value),
			})
		}
	}()
}

func mqttHandleIncoming(m *proto.Publish) {
	topic := m.TopicName
	message := []byte(m.Payload.(proto.BytesPayload))
	switch {
	case !strings.HasPrefix(topic, "$SYS"): log.Printf("msg %s = %s r: %v", topic, message, m.Header.Retain)
	}
	
}

func Publish(key string, value []byte) {
	log.Printf("P %s => %s", key, value)
	mqttClient.Publish(&proto.Publish{
		// Header:    proto.Header{Retain: true},
		TopicName: key,
		Payload:   proto.BytesPayload(value),
	})
}

