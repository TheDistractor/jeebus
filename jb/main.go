// The JeeBus server, with messaging, data storage, and a web server.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/chimera/rs232"
	"github.com/jcw/jeebus"
	"github.com/jeffallen/mqtt"
	"github.com/syndtr/goleveldb/leveldb"
	// "github.com/syndtr/goleveldb/leveldb/opt"
)

var regClient, dbClient, ifClient, wsClient, rdClient, svClient jeebus.Client

type TextMessage struct {
	Text string `json:"text"`
}

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
		for m := range jeebus.ConnectToServer(topics) {
			retain := ""
			if m.R {
				retain = "(retain)"
			}
			log.Println(m.T, string(m.P), retain)
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

	if to == "" {
		to = from + "~" // FIXME this assumes all key chars are less than "~"
	}

	// get and print all the key/value pairs from the database
	iter := db.NewIterator(nil)
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

	log.Println("setting up Lua")
	setupLua()

	log.Println("starting MQTT server")
	sock, err := net.Listen("tcp", ":1883")
	check(err)
	svr := mqtt.NewServer(sock)
	svr.Start()
	// <-svr.Done
	log.Println("MQTT server is running")

	regClient.Connect("@")
	regClient.Register("#", &RegistryService{})

	dbClient.Connect("")
	dbClient.Register("#", &DatabaseService{db})

	ifClient.Connect("if")
	wsClient.Connect("ws")

	rdClient.Connect("rd")
	rdClient.Register("blinker/#", new(BlinkerDecodeService))

	svClient.Connect("sv")
	svClient.Register("blinker/#", new(BlinkerEncodeService))

	jeebus.Publish("/admin/started", time.Now().Format(time.RFC822Z))

	log.Println("starting web server on ", port)
	http.Handle("/", http.FileServer(http.Dir("./app")))
	http.Handle("/ws", websocket.Handler(sockServer))
	log.Fatal(http.ListenAndServe(port, nil))
}

type RegistryService map[string]map[string]byte

func (s *RegistryService) Handle(m *jeebus.Message) {
	split := strings.SplitN(m.T, "/", 2)
	var arg string
	err := json.Unmarshal(m.P, &arg)
	check(err)
	log.Printf("REG %s = %s", m.T, arg)

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

	log.Printf("  %+v", *s)
}

type DatabaseService struct {
	db *leveldb.DB // TODO can't this struct nesting be avoided, somehow?
}

func (s *DatabaseService) Handle(m *jeebus.Message) {
	s.db.Put([]byte(m.T), m.P, nil)
	millis := time.Now().UnixNano() / 1000000
	s.db.Put([]byte(fmt.Sprintf("hist/%s/%d", m.T, millis)), m.P, nil)
}

type SerialInterfaceService struct {
	serial *rs232.Port // TODO can't this struct nesting be avoided, somehow?
}

func (s *SerialInterfaceService) Handle(m *jeebus.Message) {
	var arg struct{ Text string }
	err := json.Unmarshal(m.P, &arg)
	check(err)
	s.serial.Write([]byte(arg.Text))
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
	log.Println("serial ready:", name)

	ifClient.Register(name, &SerialInterfaceService{serial})

	for scanner.Scan() {
		jeebus.Publish("rd/"+name, &TextMessage{scanner.Text()})
	}

	ifClient.Unregister(name)
}

type WebsocketService struct {
	ws *websocket.Conn // TODO can't this struct nesting be avoided, somehow?
}

func (s *WebsocketService) Handle(m *jeebus.Message) {
	err := websocket.Message.Send(s.ws, string(m.P))
	check(err)
}

func sockServer(ws *websocket.Conn) {
	defer ws.Close()

	tag := ws.Request().Header.Get("Sec-Websocket-Protocol")
	name := tag + "/ip-" + ws.Request().RemoteAddr
	wsClient.Register(name, &WebsocketService{ws})

	for {
		var msg json.RawMessage
		err := websocket.JSON.Receive(ws, &msg)
		if err == io.EOF {
			break
		}
		check(err)
		jeebus.Publish("sv/"+name, msg)
	}

	wsClient.Unregister(name)
}

type BlinkerDecodeService int

func (s *BlinkerDecodeService) Handle(m *jeebus.Message) {
	var cmd struct{ Text string }
	err := json.Unmarshal(m.P, &cmd)
	check(err)
	num, err := strconv.Atoi(cmd.Text[1:])
	check(err)
	// TODO this is hard-coded, should probably be a lookup table set via pub's
	// TODO yuck, would be a lot cleaner in dynamically-typed Lua, etc
	var x interface{}
	switch cmd.Text[0] {
	case 'C':
		x = map[string]int{"count": num}
	case 'G':
		x = map[string]bool{"green": num != 0}
	case 'R':
		x = map[string]bool{"red": num != 0}
	}
	jeebus.Publish("ws/blinker", x)
}

type BlinkerEncodeService int

func (s *BlinkerEncodeService) Handle(m *jeebus.Message) {
	var arg struct{ Button, Value int }
	err := json.Unmarshal(m.P, &arg)
	check(err)
	// TODO this is hard-coded, should probably be a lookup table set via pub's
	msg := fmt.Sprintf("L%d%d", arg.Button, arg.Value)
	jeebus.Publish("if/blinker", &TextMessage{msg})
}
