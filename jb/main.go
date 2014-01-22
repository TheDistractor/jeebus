package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"code.google.com/p/go.net/websocket"
	"github.com/chimera/rs232"
	"github.com/jcw/jeebus"
	"github.com/jeffallen/mqtt"
	"github.com/syndtr/goleveldb/leveldb"
	// "github.com/syndtr/goleveldb/leveldb/opt"
)

var (
	openWebSockets map[string]*websocket.Conn
	mqttClient     *mqtt.ClientConn // TODO get rid of this
	dataStore      *leveldb.DB
)

func main() {
	if len(os.Args) <= 1 {
		log.Fatalf("usage: jb <cmd> ... (try 'jb run')")
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
		for m := range jeebus.ListenToServer(topics) {
			log.Println(m.T, string(m.P.([]byte)), m.R)
		}

	case "serial":
		if len(os.Args) <= 2 {
			log.Fatalf("usage: jb serial <dev> ?baud? ?tag?")
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
		feed := jeebus.ListenToServer("if/serial")

		log.Println("opening serial port", dev)
		serial := serialConnect(dev, nbaud, tag)

		for m := range feed {
			log.Printf("Ser: %s", m.P.([]byte))
			serial.Write(m.P.([]byte))
		}

	default:
		log.Fatal("unknown sub-command: jb ", os.Args[1], " ...")
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
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
			jeebus.PubChan <- &jeebus.Message{T: serKey, P: line}
		}

		log.Printf("no more data on: %s", dev)
	}()

	return ser
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

func sockServer(ws *websocket.Conn) {
	defer ws.Close()
	client := ws.Request().RemoteAddr
	openWebSockets[client] = ws
	log.Println("Client connected:", client)

	for {
		// var msg string
		var any interface{}
		err := websocket.JSON.Receive(ws, &any)
		if err != nil {
			log.Print(err)
			break
		}
		fmt.Printf("ws got: %#v\n", any)
		// jeebus.Publish("ws/"+client, []byte(msg))
	}

	log.Println("Client disconnected:", client)
	delete(openWebSockets, client)
}
