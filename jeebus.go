package jeebus

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
	// "github.com/syndtr/goleveldb/leveldb/opt"
)

var (
	openWebSockets map[string]*websocket.Conn
	mqttClient     *mqtt.ClientConn // TODO get rid of this
	dataStore      *leveldb.DB
	PubChan        chan *Message
)

type Message struct {
	T string      // topic
	P interface{} // payload
	R bool        // retain
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ListenToServer(topic string) chan Message {
	conn, err := net.Dial("tcp", "localhost:1883")
	check(err)

	mqttClient = mqtt.NewClientConn(conn)
	err2 := mqttClient.Connect("", "")
	check(err2)

	mqttClient.Subscribe([]proto.TopicQos{
		{Topic: topic, Qos: proto.QosAtMostOnce},
	})

	// set up a channel to publish through
	PubChan = make(chan *Message)
	go func() {
		for msg := range PubChan {
			// log.Printf("C %s => %v", msg.T, msg.P)
			value, err := json.Marshal(msg.P)
			check(err)
			mqttClient.Publish(&proto.Publish{
				Header:    proto.Header{Retain: msg.R},
				TopicName: msg.T,
				Payload:   proto.BytesPayload(value),
			})
		}
	}()

	listenChan := make(chan Message)
	go func() {
		for m := range mqttClient.Incoming {
			listenChan <- Message{
				T: m.TopicName,
				P: []byte(m.Payload.(proto.BytesPayload)),
				R: m.Header.Retain,
			}
		}
		log.Println("server connection lost")
		close(listenChan)
	}()

	return listenChan
}

func SubCommand(cmdName string) string {
	// TODO figure out how to use the "flag" package with sub-commands
	if len(os.Args) <= 1 {
		log.Fatalf("usage: %s <cmd> ... (try '%s run')", cmdName, cmdName)
	}

	switch os.Args[1] {

	case "dump":
		switch len(os.Args) {
		case 2:
			dumpDatabase("", "")
		case 3:
			dumpDatabase(os.Args[2], "")
		case 4:
			dumpDatabase(os.Args[2], os.Args[3])
		}

	case "run":
		port := ":3333"
		if len(os.Args) > 2 {
			port = os.Args[2]
		}
		startAllServers(port)

	case "see":
		topics := "#"
		if len(os.Args) > 2 {
			topics = os.Args[2]
		}
		for m := range ListenToServer(topics) {
			log.Println(m.T, string(m.P.([]byte)), m.R)
		}

	case "serial":
		if len(os.Args) <= 2 {
			log.Fatalf("usage: %s serial <dev> ?baud? ?tag?", cmdName)
		}
		dev, baud, tag := os.Args[2], "57600", ""
		if len(os.Args) > 3 {
			baud = os.Args[3]
		}
		if len(os.Args) > 4 {
			tag = os.Args[4]
		}
		nbaud, err := strconv.Atoi(baud)
		check(err)
		feed := ListenToServer("if/serial")

		log.Println("opening serial port", dev)
		serial := serialConnect(dev, nbaud, tag)

		for m := range feed {
			log.Printf("Ser: %s", m.P.([]byte))
			serial.Write(m.P.([]byte))
		}

	default:
		return os.Args[1]
	}

	os.Exit(0) // sub-command has been processed, normal exit
	return ""  // never reached
}

func dumpDatabase(from, to string) {
	// o := &opt.Options{ ErrorIfMissing: true }
	db, err := leveldb.OpenFile("./storage", nil)
	check(err)
	dataStore = db

	if to == "" {
		to = from + "~" // FIXME this assumes all key chars are less than "~"
	}

	// get and print all the key/value pairs from the database
	iter := dataStore.NewIterator(nil)
	iter.Seek([]byte(from))
	for iter.Valid() {
		fmt.Printf("%s = %s\n", iter.Key(), iter.Value())
		if !iter.Next() || string(iter.Key()) > to {
			break
		}
	}
	iter.Release()
}

func startAllServers(port string) {
	openWebSockets = make(map[string]*websocket.Conn)

	log.Println("opening database")
	db, err := leveldb.OpenFile("./storage", nil)
	check(err)
	dataStore = db

	log.Println("setting up Lua")
	setupLua()

	log.Println("starting MQTT server")
	startMqttServer()
	log.Println("MQTT server is running")

	// set up a web server to handle static files and websockets
	http.Handle("/", http.FileServer(http.Dir("./app")))
	http.Handle("/ws", websocket.Handler(sockServer))
	log.Println("web server started on ", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func startMqttServer() {
	port, err := net.Listen("tcp", ":1883")
	check(err)
	svr := mqtt.NewServer(port)
	svr.Start()
	// <-svr.Done

	feed := ListenToServer("#")

	publish("st/admin/started", []byte(time.Now().Format(time.RFC822Z)))

	go func() {
		for m := range feed {
			mqttDispatch(&m)
		}
	}()
}

func mqttDispatch(m *Message) {
	topic := m.T
	message := m.P.([]byte)

	// FIXME can't work: retain flag is not published to subscribers!
	//	solving this will require a modified mqtt package
	// if m.Header.Retain {
	// 	store("mqtt/"+topic, message, nil)
	// }

	switch topic[:3] {

	// st/key... -> current state, stored as key and with timestamp
	case "st/":
		key := topic[3:]
		store(key, message)
		millis := time.Now().UnixNano() / 1000000
		store(fmt.Sprintf("hist/%s/%d", key, millis), message)

	// db/... -> database requests, value is reply topic
	case "db/":
		if strings.HasPrefix(topic, "db/get/") {
			value := fetch(topic[7:])
			publish(string(message), value)
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
		publish("if/serial", []byte(cmd))
	}
}

// TODO get rid of this, use PubChan and add support for raw sending
func publish(key string, value []byte) {
	// log.Printf("P %s => %s", key, value)
	mqttClient.Publish(&proto.Publish{
		// Header:    proto.Header{Retain: true},
		TopicName: key,
		Payload:   proto.BytesPayload(value),
	})
}

func fetch(key string) []byte {
	value, err := dataStore.Get([]byte(key), nil)
	if err != nil {
		log.Println(err)
	}
	return value
}

func store(key string, value []byte) {
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
			PubChan <- &Message{T: serKey, P: line}
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
		publish("ws/"+client, []byte(msg))
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
