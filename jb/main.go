package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"code.google.com/p/go.net/websocket"
	"github.com/chimera/rs232"
	"github.com/jcw/jeebus"
	"github.com/syndtr/goleveldb/leveldb"
	// "github.com/syndtr/goleveldb/leveldb/opt"
)

var (
	openWebSockets map[string]*websocket.Conn
	dataStore      *leveldb.DB
	pubChan        chan *jeebus.Message
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
		_, sub := jeebus.ConnectToServer(topics)
		for m := range sub {
			log.Println(m.T, string(m.P), m.R)
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
		log.Println("opening serial port", dev)
		<-serialConnect(dev, nbaud, tag)

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

func serialConnect(dev string, baudrate int, tag string) (done chan byte) {
	// open the serial port in 8N1 mode
	serial, err := rs232.Open(dev, rs232.Options{
		BitRate: uint32(baudrate), DataBits: 8, StopBits: 1,
	})
	check(err)

	port := strings.TrimPrefix(dev, "/dev/")
	port = strings.Replace(port, "tty.usbserial-", "usb-", 1)

	done = make(chan byte)

	go func() {
		scanner := bufio.NewScanner(serial)

		// flush all old data from the serial port while looking for a tag
		if tag == "" {
			log.Println("waiting for serial")
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "[") && strings.Contains(line, "]") {
					tag = line[1:strings.IndexAny(line, ".]")]
					log.Println("serial started:", tag)
					break
				}
			}
		}

		serKey := "if/serial/" + tag + "/" + port
		// TODO listen to all for now, until server can broadcast
		// pub, sub := jeebus.ConnectToServer(":" + serKey)
		pub, sub := jeebus.ConnectToServer(":if/serial/" + tag + "/#")

		// send out published commands
		go func() {
			defer serial.Close()
			for m := range sub {
				log.Printf("Ser: %s", m.P)
				serial.Write(m.P)
			}
		}()

		// publish incoming data as a JSON string
		for scanner.Scan() {
			msg, err := json.Marshal(scanner.Text())
			check(err)
			pub <- &jeebus.Message{T: serKey, P: msg}
		}

		log.Printf("no more data on: %s", port)
		done <- 1
	}()

	return
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
	pubChan = startMqttServer()
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
	subProto := ws.Request().Header.Get("Sec-Websocket-Protocol")
	log.Println("ws connect", subProto, client)

	for {
		var msg []byte
		err := websocket.Message.Receive(ws, &msg)
		if err == io.EOF {
			break
		}
		check(err)
		fmt.Printf("ws got: %s\n", msg)
		pubChan <- &jeebus.Message{T: "ws/" + client, P: msg}
	}

	log.Println("ws disconnect", subProto, client)
	delete(openWebSockets, client)
}

func sendToAllWebSockets(m []byte) {
	for _, ws := range openWebSockets {
		err := websocket.Message.Send(ws, string(m))
		check(err)
	}
}
