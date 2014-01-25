// The JeeBus server, with messaging, data storage, and a web server.
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/chimera/rs232"
	"github.com/jcw/jeebus"
	"github.com/syndtr/goleveldb/leveldb"
	// "github.com/syndtr/goleveldb/leveldb/opt"
)

var (
	dataStore *leveldb.DB
	regClient jeebus.Client
	dbClient  jeebus.Client
	ifClient  jeebus.Client
	wsClient  jeebus.Client
	svClient  jeebus.Client
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
		ifClient.Connect("if")
		serialConnect(dev, nbaud, tag)

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

func startAllServers(port string) {
	log.Println("opening database")
	db, err := leveldb.OpenFile("./storage", nil)
	check(err)
	dataStore = db

	log.Println("setting up Lua")
	setupLua()

	log.Println("starting MQTT server")
	startMqttServer()
	log.Println("MQTT server is running")

	regClient.Connect("@")
	regClient.Register("#", &RegistryService{})

	dbClient.Connect("")
	dbClient.Register("#", new(DatabaseService))

	ifClient.Connect("if")
	wsClient.Connect("ws")

	svClient.Connect("sv")
	svClient.Register("blinker", new(BlinkerService))

	// TODO should use new "/..." style
	regClient.Publish("st/admin/started", time.Now().Format(time.RFC822Z))

	log.Println("starting web server on ", port)
	http.Handle("/", http.FileServer(http.Dir("./app")))
	http.Handle("/ws", websocket.Handler(sockServer))
	log.Fatal(http.ListenAndServe(port, nil))
}

type RegistryService map[string]map[string]byte

func (s *RegistryService) Handle(tail string, value interface{}) {
	log.Printf("@ '%s', value %#v (%T)", tail, value, value)
	split := strings.SplitN(tail, "/", 2)
	arg := value.(string)

	switch split[0] {
	case "connect":
		(*s)[arg] = make(map[string]byte)
	case "disconnect":
		delete(*s, arg)
	case "register":
		(*s)[split[1]][arg] = 1
	case "unregister":
		delete((*s)[split[1]], arg)
	}

	log.Printf("registry %v", *s)
}

type DatabaseService int

func (s *DatabaseService) Handle(tail string, value interface{}) {
	log.Printf("DB '%s', value %#v (%T)", tail, value, value)
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

type SerialInterfaceService struct {
	serial *rs232.Port // TODO can't this struct nesting be avoided, somehow?
}

func (s *SerialInterfaceService) Handle(tail string, value interface{}) {
	s.serial.Write([]byte(value.(string))) // TODO yuck, messy cast
}

func serialConnect(dev string, baudrate int, tag string) {
	// open the serial port in 8N1 mode
	serial, err := rs232.Open(dev, rs232.Options{
		BitRate: uint32(baudrate), DataBits: 8, StopBits: 1,
	})
	check(err)

	scanner := bufio.NewScanner(serial)

	// flush all old data from the serial port while looking for a tag
	if tag == "" {
		log.Println("waiting for serial")
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "[") && strings.Contains(line, "]") {
				tag = line[1:strings.IndexAny(line, ".]")]
				break
			}
		}
	}

	port := strings.TrimPrefix(dev, "/dev/")
	name := tag + "/" + strings.Replace(port, "tty.usbserial-", "usb-", 1)

	ifClient.Register(name, &SerialInterfaceService{serial})

	for scanner.Scan() {
		// FIXME confused about broadcasts, probably need a "#" in register?
		// ifClient.Publish("sv/" + name, scanner.Text())
		ifClient.Publish("sv/"+tag, scanner.Text())
	}

	ifClient.Unregister(name)
}

type WebsocketService struct {
	ws *websocket.Conn // TODO can't this struct nesting be avoided, somehow?
}

func (s *WebsocketService) Handle(tail string, value interface{}) {
	err := websocket.JSON.Send(s.ws, value)
	check(err)
}

func sockServer(ws *websocket.Conn) {
	defer ws.Close()

	name := ws.Request().Header.Get("Sec-Websocket-Protocol")
	name += "/" + ws.Request().RemoteAddr
	wsClient.Register(name, &WebsocketService{ws})

	for {
		var any []string
		err := websocket.JSON.Receive(ws, &any)
		if err == io.EOF {
			break
		}
		check(err)
		wsClient.Publish(any[0], any[1])
	}

	wsClient.Unregister(name)
}

type BlinkerService int

func (s *BlinkerService) Handle(tail string, value interface{}) {
	// TODO this is hard-coded, should probably be a lookup table set via pub's
	svClient.Publish("ws/blinker", value)
}
