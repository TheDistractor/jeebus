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
		case "see":
			seeCmd()
		case "serial":
			serialCmd()
		default:
			server()
		}
	} else {
		server()
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func listenToServer(topic string) chan *proto.Publish {
	conn, err := net.Dial("tcp", "localhost:1883")
	check(err)

	mqttClient = mqtt.NewClientConn(conn)
	err2 := mqttClient.Connect("", "")
	check(err2)

	mqttClient.Subscribe([]proto.TopicQos{
		{Topic: topic, Qos: proto.QosAtMostOnce},
	})

	// set up a channel to publish through
	busPubChan = make(chan *BusMessage)

	go func() {
		for msg := range busPubChan {
			// log.Printf("C %s => %v", msg.T, msg.M)
			value, err := json.Marshal(msg.M)
			check(err)
			mqttClient.Publish(&proto.Publish{
				Header:    proto.Header{Retain: msg.R},
				TopicName: msg.T,
				Payload:   proto.BytesPayload(value),
			})
		}
	}()

	return mqttClient.Incoming
}

func seeCmd() {
	for m := range listenToServer("#") {
		topic := m.TopicName
		message := []byte(m.Payload.(proto.BytesPayload))
		retain := ""
		if m.Header.Retain {
			retain = " (retain)"
		}
		log.Println(topic+retain, "=", string(message))
	}
}

func serialCmd() {
	if len(os.Args) < 4 {
		log.Fatal("usage: jeebus serial <dev> <baud> ?tag?")
	}
	dev, sbaud, tag := os.Args[2], os.Args[3], ""
	if len(os.Args) > 4 {
		tag = os.Args[4]
	}
	nbaud, err := strconv.Atoi(sbaud)
	check(err)

	feed := listenToServer("if/serial")

	log.Println("opening serial port", dev)
	serialPort = serialConnect(dev, nbaud, tag)

	for m := range feed {
		message := []byte(m.Payload.(proto.BytesPayload))
		log.Printf("Ser: %s", message)
		serialPort.Write(message)
	}
}

func server() {
	openWebSockets = make(map[string]*websocket.Conn)

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

	// set up a web server to handle static files and websockets
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.Handle("/ws", websocket.Handler(sockServer))
	log.Println("web server is listening on port 3333")
	log.Fatal(http.ListenAndServe(":3333", nil))
}

func startMqttServer() {
	port, err := net.Listen("tcp", ":1883")
	check(err)
	svr := mqtt.NewServer(port)
	svr.Start()
	// <-svr.Done

	feed := listenToServer("#")

	Publish("st/admin/started", []byte(time.Now().Format(time.RFC822Z)))

	go func() {
		for m := range feed {
			mqttDispatch(m)
		}
	}()
}

func mqttDispatch(m *proto.Publish) {
	topic := m.TopicName
	message := []byte(m.Payload.(proto.BytesPayload))

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
		// log.Printf("got %s", message)
		err := json.Unmarshal(message, &any)
		check(err)
		// send as L<n><m> to the serial port
		cmd := fmt.Sprintf("L%.0f%.0f", any[0], any[1])
		Publish("if/serial", []byte(cmd))
	}
}

// TODO get rid of this, use busPubChan and add support for raw sending
func Publish(key string, value []byte) {
	// log.Printf("P %s => %s", key, value)
	mqttClient.Publish(&proto.Publish{
		// Header:    proto.Header{Retain: true},
		TopicName: key,
		Payload:   proto.BytesPayload(value),
	})
}

func openDatabase(dbname string) {
	db, err := leveldb.OpenFile(dbname, nil)
	check(err)
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

func serialConnect(dev string, baud int, tag string) *rs232.Port {
	// open the serial port
	options := rs232.Options{
		BitRate:  uint32(baud),
		DataBits: 8,
		StopBits: 1,
	}
	ser, err := rs232.Open(dev, options)
	check(err)

	// turn incoming data into a channel of text lines
	inputLines := make(chan string)

	go func() {
		scanner := bufio.NewScanner(ser)
		for scanner.Scan() {
			inputLines <- scanner.Text()
		}
		log.Printf("serial port disconnect: %s", dev)
		close(inputLines)
	}()

	// publish incoming data
	go func() {
		// flush all old data from the serial port
		if tag == "" {
			log.Println("waiting for serial")
			for line := range inputLines {
				if strings.HasPrefix(line, "[") && strings.Contains(line, "]") {
					tag = line[1:strings.IndexAny(line, ".]")]
					break
				}
			}
			log.Println("serial started:", tag)
		}

		serKey := "if/serial/" + tag + "/" + strings.TrimPrefix(dev, "/dev/")
		for line := range inputLines {
			busPubChan <- &BusMessage{T: serKey, M: line}
		}

		log.Printf("no more data on: %s", dev)
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
