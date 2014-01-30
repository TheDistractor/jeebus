// The JeeBus server, with messaging, data storage, and a web server.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
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

var regClient, dbClient, ifClient, wsClient, rdClient, svClient *jeebus.Client

var db *leveldb.DB

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
			log.Println(m.T, string(m.P))
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
		ifClient = jeebus.NewClient("if")
		serialConnect(dev, nbaud, tag)

	case "pub":
		if len(os.Args) < 3 {
			log.Fatalf("usage: jb pub <key> ?<jsonval>?")
		}
		var value string
		if len(os.Args) > 3 {
			value = os.Args[3]
		}
		sub := jeebus.ConnectToServer("?") // TODO nonsense topic
		jeebus.Publish(os.Args[2], []byte(value))
		// TODO need to close gracefully, and not too soon!
		time.Sleep(10 * time.Millisecond)
		close(sub)

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
	var err error

	log.Println("opening database")
	db, err = leveldb.OpenFile("./storage", nil)
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

	regClient = jeebus.NewClient("@")
	regClient.Register("#", &RegistryService{})

	dbClient = jeebus.NewClient("")
	dbClient.Register("#", &DatabaseService{db})

	ifClient = jeebus.NewClient("if")
	wsClient = jeebus.NewClient("ws")

	rdClient = jeebus.NewClient("rd")
	rdClient.Register("blinker/#", new(BlinkerDecodeService))
	rdClient.Register("#", new(LoggerService))

	svClient = jeebus.NewClient("sv")
	svClient.Register("blinker/#", new(BlinkerEncodeService))
	svClient.Register("lua/#", new(LuaDispatchService))

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
}

type DatabaseService struct {
	db *leveldb.DB // TODO can't this struct nesting be avoided, somehow?
}

func (s *DatabaseService) Handle(m *jeebus.Message) {
	if len(m.P) > 0 {
		s.db.Put([]byte("/"+m.T), m.P, nil)
		millis := time.Now().UnixNano() / 1000000
		s.db.Put([]byte(fmt.Sprintf("hist/%s/%d", m.T, millis)), m.P, nil)
	} else {
		s.db.Delete([]byte("/"+m.T), nil)
		// TODO decide what to do with deletions w.r.t. the historical data
		//  record the deletion? delete it as well? sweep and clean up later?
	}
}

type SerialInterfaceService struct {
	serial *rs232.Port // TODO can't this struct nesting be avoided, somehow?
}

func (s *SerialInterfaceService) Handle(m *jeebus.Message) {
	s.serial.Write([]byte(m.Get("text")))
}

func serialConnect(port string, baudrate int, tag string) {
	// open the serial port in 8N1 mode
	serial, err := rs232.Open(port, rs232.Options{
		BitRate: uint32(baudrate), DataBits: 8, StopBits: 1,
	})
	check(err)

	scanner := bufio.NewScanner(serial)

	var input struct {
		Text string `json:"text"`
		Time int64  `json:"time"`
	}

	// flush all old data from the serial port while looking for a tag
	if tag == "" {
		log.Println("waiting for serial")
		for scanner.Scan() {
			input.Time = time.Now().UTC().UnixNano() / 1000000
			input.Text = scanner.Text()
			if strings.HasPrefix(input.Text, "[") &&
				strings.Contains(input.Text, "]") {
				tag = input.Text[1:strings.IndexAny(input.Text, ".]")]
				break
			}
		}
	}

	dev := strings.TrimPrefix(port, "/dev/")
	dev = strings.Replace(dev, "tty.usbserial-", "usb-", 1)
	name := tag + "/" + dev
	log.Println("serial ready:", name)

	ifClient.Register(name, &SerialInterfaceService{serial})

	// store the tag line for this device
	attachMsg := map[string]string{"text": input.Text, "tag": tag}
	jeebus.Publish("/attach/"+dev, attachMsg)

	// send the tag line (if present), then send out whatever comes in
	if input.Text != "" {
		jeebus.Publish("rd/"+name, &input)
	}
	for scanner.Scan() {
		input.Time = time.Now().UTC().UnixNano() / 1000000
		input.Text = scanner.Text()
		jeebus.Publish("rd/"+name, &input)
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
	defer wsClient.Unregister(name)

	for {
		var msg json.RawMessage
		err := websocket.JSON.Receive(ws, &msg)
		if err == io.EOF {
			break
		}
		check(err)
		// figure out the structure of incoming JSON data to decide what to do
		switch msg[0] {
		// case 'n': // null
		// 	log.Println("shutdown requested from", name)
		// 	os.Exit(0)
		case '"':
			// show incoming JSON strings on JB's stdout for debugging purposes
			var text string
			err := json.Unmarshal(msg, &text)
			check(err)
			log.Printf("%s (%s)", text, name) // send to JB server's stdout
		case '[':
			// JSON array: either an MQTT publish request, or an RPC request
			var args []json.RawMessage
			err := json.Unmarshal(msg, &args)
			check(err)
			var topic string
			if json.Unmarshal(args[0], &topic) == nil {
				// it's an MQTT publish request
				log.Println("TOPIC", topic, args)
				jeebus.Publish(topic, args[1])
			} else {
				// it's an RPC request of the form (rpcId, req string, args...]
				var any []interface{}
				err := json.Unmarshal(msg, &any)
				check(err)
				log.Printf("RPC %.100v (%s)", any, name)
				// ptocess the RPC request, returns either a value or an error
				result, err := processRpcRequest(any[1].(string), any[2:])
				// convert errors to strings to send them through JSON
				var emsg interface{}
				if err != nil {
					emsg = err.Error()
				}
				reply := []interface{}{any[0], result, emsg}
				log.Printf(" -> %.100v (%s)", reply, name)
				msg, err := json.Marshal(reply)
				check(err)
				err = websocket.Message.Send(ws, string(msg))
				check(err)
			}
		default:
			// everything else (i.e. a JSON object) becomes an MQTT service req
			jeebus.Publish("sv/"+name, msg)
		}
	}
}

func processRpcRequest(cmd string, args []interface{}) (r interface{}, e error) {
	switch cmd {
	case "echo":
		return args, nil
	case "db-keys":
		return dbKeys(args[0].(string)), nil
	case "db-get":
		v, e := db.Get([]byte(args[0].(string)), nil) // TODO yuck...
		return string(v), e
	}
	return nil, errors.New("RPC not found: " + cmd)
}

func dbKeys(prefix string) []string {
	// TODO decide whether this key logic is the most useful & least confusing
	// TODO should use skips and reverse iterators once the db gets larger!
	from, to, skip := []byte(prefix), []byte(prefix+"~"), len(prefix)
	// from, to, skip := []byte(prefix+"/"), []byte(prefix+"/~"), len(prefix)+1
	result := []string{}
	prev := []byte("/") // impossible value, this never matches actual results

	iter := db.NewIterator(nil)
	defer iter.Release()

	iter.Seek(from)
	for iter.Valid() {
		k := iter.Key()
		// fmt.Printf(" -> %s = %s\n", k, iter.Value())
		if !iter.Next() || bytes.Compare(k, to) > 0 {
			break
		}
		i := bytes.IndexRune(k[skip:], '/') + skip
		if i < skip {
			i = len(k)
		}
		// fmt.Printf(" DK %d %d %d %s %s\n", skip, len(prev), i, prev, k)
		if !bytes.Equal(prev, k[skip:i]) {
			// TODO need to make a copy of the key, since it's owned by iter
			prev = make([]byte, i-skip)
			copy(prev, k[skip:i])
			// fmt.Printf("ADD %s\n", prev)
			result = append(result, string(prev))
		}
	}
	return result
}

type BlinkerDecodeService int

func (s *BlinkerDecodeService) Handle(m *jeebus.Message) {
	text := m.Get("text")
	num, _ := strconv.Atoi(text[1:])
	// TODO this is hard-coded, should probably be a lookup table set via pub's
	msg := make(map[string]interface{})
	switch text[0] {
	case 'C':
		msg["count"] = num
	case 'G':
		msg["green"] = num != 0
	case 'R':
		msg["red"] = num != 0
	}
	jeebus.Publish("ws/blinker", msg)
}

type BlinkerEncodeService int

func (s *BlinkerEncodeService) Handle(m *jeebus.Message) {
	// TODO this is hard-coded, should probably be a lookup table set via pub's
	msg := map[string]interface{}{
		"text": fmt.Sprintf("L%d%d", m.GetInt("button"), m.GetInt("value")),
	}
	jeebus.Publish("if/blinker", msg)
}

type LoggerService struct {
	fd *os.File // TODO can't this struct nesting be avoided, somehow?
}

// LOGGER_PREFIX is where log files get created. While this directory exists,
// the logger will store new files in it and append log items. Note that it is
// perfectly ok to create or remove this directory while the logger is running.
const LOGGER_PREFIX = "./logger/"

func (s *LoggerService) Handle(msg *jeebus.Message) {
	// automatic enabling/disabling of the logger, based on presence of dir
	_, err := os.Stat(LOGGER_PREFIX)
	if err != nil {
		if s.fd != nil {
			log.Println("logger stopped")
			s.fd.Close()
			s.fd = nil
		}
		return
	}
	if s.fd == nil {
		log.Println("logger started")
	}
	// figure out name of logfile based on UTC date, with daily rotation
	now := time.Now().UTC()
	datePath := dateFilename(now)
	if s.fd == nil || datePath != s.fd.Name() {
		if s.fd != nil {
			s.fd.Close()
		}
		mode := os.O_WRONLY | os.O_APPEND | os.O_CREATE
		fd, err := os.OpenFile(datePath, mode, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		s.fd = fd
	}
	// append a new log entry, here is an example of the format used:
	// 	L 01:02:03.537 usb-A40117UK OK 9 25 54 66 235 61 210 226 33 19
	hour, min, sec := now.Clock()
	port := strings.SplitN(msg.T, "/", 2)[1] // skip the service name
	line := fmt.Sprintf("L %02d:%02d:%02d.%03d %s %s\n",
		hour, min, sec, now.Nanosecond()/1000000, port, msg.Get("text"))
	s.fd.WriteString(line)
}

func dateFilename(now time.Time) string {
	year, month, day := now.Date()
	path := fmt.Sprintf("%s%d", LOGGER_PREFIX, year)
	os.MkdirAll(path, os.ModePerm)
	// e.g. "./logger/2014/20140122.txt"
	return fmt.Sprintf("%s/%d.txt", path, (year*100+int(month))*100+day)
}
