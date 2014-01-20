package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/aarzilli/golua/lua"
	"github.com/chimera/rs232"
	proto "github.com/huin/mqtt"
	"github.com/jeffallen/mqtt"
	"github.com/stevedonovan/luar"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	openWebSockets map[string]*websocket.Conn
	serialPort     *rs232.Port
	mqttClient     *mqtt.ClientConn
	dataStore      *leveldb.DB
	busPubChan     chan *BusMessage
)

type BusMessage struct {
	T string      // topic
	M interface{} // message
	R bool        // retain
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "activity":
			activity()
		default:
			server()
		}
	} else {
		server()
	}
}

func activity() {
	conn, _ := net.Dial("tcp", "localhost:1883")
	mqttClient = mqtt.NewClientConn(conn)
	mqttClient.Connect("", "")

	mqttClient.Subscribe([]proto.TopicQos{
		{Topic: "#", Qos: proto.QosAtMostOnce},
	})

	for m := range mqttClient.Incoming {
		topic := m.TopicName
		message := []byte(m.Payload.(proto.BytesPayload))
		retain := ""
		if m.Header.Retain {
			retain = " (retain)"
		}
		log.Println(topic+retain, "=", string(message))
	}
}

func server() {
	openWebSockets = make(map[string]*websocket.Conn)
	busPubChan = make(chan *BusMessage)

	log.Println("opening database")
	openDatabase("./storage")

	// get and print all the key/value pairs from the database
	iter := dataStore.NewIterator(nil)
	for iter.Next() {
		log.Printf("key: %s, value: %s\n", iter.Key(), iter.Value())
	}
	iter.Release()

	log.Println("setting up Lua")
	setupLua()

	log.Println("starting MQTT server")
	startMqttServer()
	log.Println("MQTT server is running")

	// passing serial port as first arg will override the default
	dev := "/dev/tty.usbserial-A40115A2"
	if len(os.Args) > 1 {
		dev = os.Args[1]
	}
	log.Println("opening serial port", dev)
	serialPort = serialConnect(dev)

	// set up a web server to handle static files and websockets
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.Handle("/ws", websocket.Handler(sockServer))
	log.Println("web server is listening on port 3333")
	log.Fatal(http.ListenAndServe(":3333", nil))
}

func startMqttServer() {
	ready := make(chan bool)
	go func() {
		port, err := net.Listen("tcp", ":1883")
		if err != nil {
			log.Fatal("listen: ", err)
		}
		svr := mqtt.NewServer(port)
		svr.Start()
		ready <- true

		conn, _ := net.Dial("tcp", "localhost:1883")
		mqttClient = mqtt.NewClientConn(conn)
		// mqttClient.Dump = true
		mqttClient.Connect("", "")
		Publish("st/admin/started", []byte(time.Now().Format(time.RFC822Z)))

		mqttClient.Subscribe([]proto.TopicQos{
			{Topic: "#", Qos: proto.QosAtMostOnce},
		})

		for m := range mqttClient.Incoming {
			mqttDispatch(m)
		}
		// <-svr.Done
	}()

	// resume here only when the MQTT server has actually been started
	<-ready

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

func mqttDispatch(m *proto.Publish) {
	topic := m.TopicName
	message := []byte(m.Payload.(proto.BytesPayload))
	// log.Printf("msg %s = %s r: %v", topic, message, m.Header.Retain)
	// FIXME can't work: retain flag is not published to subscribers!
	//	solving this will require a modified mqtt package
	// if m.Header.Retain {
	// 	Store("mqtt/"+topic, message, nil)
	// }
	switch topic[:3] {

	// st/key... -> current state, stored as key and with timestamp
	case "st/":
		key := topic[3:]
		Store(key, message)
		millis := time.Now().UnixNano() / 1000000
		Store(fmt.Sprintf("hist/%s/%d", key, millis), message)

	// db/... -> database requests, value is reply topic
	case "db/":
		if strings.HasPrefix(topic, "db/get/") {
			value := Fetch(topic[7:])
			Publish(string(message), value)
		}

	// TODO hardcoded serial port to websocket pass-through for now
	case "if/":
		if strings.HasPrefix(topic, "if/serial/") {
			for _, ws := range openWebSockets {
				websocket.Message.Send(ws, string(message))
			}
		}
	// TODO hardcoded websocket to serial port pass-through for now
	case "ws/":
		// accept arrays of arbitrary data types
		var any []interface{}
		log.Printf("got %#v", message)
		err := json.Unmarshal(message, &any)
		if err != nil {
			log.Fatal("err?", topic, message, err)
		}
		// send as L<n><m> to the serial port
		cmd := fmt.Sprintf("L%.0f%.0f", any[0], any[1])
		serialPort.Write([]byte(cmd))
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

func openDatabase(dbname string) {
	db, err := leveldb.OpenFile(dbname, nil)
	if err != nil {
		log.Fatal(err)
	}
	dataStore = db
}

func Fetch(key string) []byte {
	value, err := dataStore.Get([]byte(key), nil)
	if err != nil {
		log.Println(err)
	}
	return value
}

func Store(key string, value []byte) {
	log.Printf("S %s => %s", key, value)
	dataStore.Put([]byte(key), value, nil)
}

func serialConnect(dev string) *rs232.Port {
	// open the serial port
	options := rs232.Options{
		BitRate:  57600,
		DataBits: 8,
		StopBits: 1,
	}
	ser, err := rs232.Open(dev, options)
	if err != nil {
		log.Fatal(err)
	}

	// turn incoming data into a channel of text lines
	inputLines := make(chan string)

	go func() {
		scanner := bufio.NewScanner(ser)
		for scanner.Scan() {
			inputLines <- scanner.Text()
		}
	}()

	// publish incoming data
	go func() {
		// flush all old data from the serial port
		log.Println("waiting for blinker to start")
		for line := range inputLines {
			if line == "[blinker]" {
				break
			}
			// TODO bail out if another sketch type is found
		}
		log.Println("blinker start detected")

		serKey := "if/serial/" + strings.TrimPrefix(dev, "/dev/")
		for line := range inputLines {
			busPubChan <- &BusMessage{T: serKey, M: line}
		}
	}()

	return ser
}

func sockServer(ws *websocket.Conn) {
	defer ws.Close()
	client := ws.Request().RemoteAddr
	openWebSockets[client] = ws
	log.Println("Client connected:", client)

	for {
		var msg string
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			log.Print(err)
			break
		}
		Publish("ws/"+client, []byte(msg))
	}

	log.Println("Client disconnected:", client)
	delete(openWebSockets, client)
}

func test(L *lua.State) int {
	fmt.Println("hello world! from go!")
	return 0
}

func test2(L *lua.State) int {
	arg := L.CheckInteger(-1)
	argfrombottom := L.CheckInteger(1)
	fmt.Print("test2 arg: ")
	fmt.Println(arg)
	fmt.Print("from bottom: ")
	fmt.Println(argfrombottom)
	return 0
}

func GoFun(args []int) (res map[string]int) {
	res = make(map[string]int)
	for i, val := range args {
		res[strconv.Itoa(i)] = val * val
	}
	return
}

const code = `
print 'here we go'
-- Lua tables auto-convert to slices
local res = GoFun {10,20,30,40}
-- the result is a map-proxy
print(res['1'],res['2'])
-- which we may explicitly convert to a table
res = luar.map2table(res)
for k,v in pairs(res) do
      print(k,v)
end
`

func setupLua() {
	L := lua.NewState()
	defer L.Close()
	L.OpenLibs()

	// L.Register("test2", test2)

	L.GetField(lua.LUA_GLOBALSINDEX, "print")
	L.PushString("Hello World!")
	L.Call(1, 0)

	L.PushGoFunction(test)
	L.PushGoFunction(test)
	L.PushGoFunction(test)
	L.PushGoFunction(test)

	L.PushGoFunction(test2)
	L.PushInteger(42)
	L.Call(1, 0)

	L.Call(0, 0)
	L.Call(0, 0)
	L.Call(0, 0)

	luar.Register(L, "", luar.Map{
		"Print": fmt.Println,
		"MSG":   "hello", // can also register constants
		"GoFun": GoFun,
	})

	err := L.DoString(code)
	fmt.Printf("error %v\n", err)
}
