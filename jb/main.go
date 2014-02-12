// The JeeBus server, with messaging, data storage, and a web server.
package main

import (
	//"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"code.google.com/p/go.net/websocket"
	//"github.com/chimera/rs232"
	"github.com/jcw/jeebus"
	"github.com/jeffallen/mqtt"
	"github.com/syndtr/goleveldb/leveldb"
	// "github.com/syndtr/goleveldb/leveldb/opt"
)

const version = "0.2-beta"

var (
	db       *leveldb.DB
	attached map[string]map[string]int // map from prefix -> tag -> refcount
	client   *jeebus.Client
)

func init() {
	log.SetFlags(log.Ltime)
	attached = make(map[string]map[string]int)
}

func main() {
	log.Println("JeeBus", version)

	if len(os.Args) <= 1 {
		log.Fatalf("usage: jb <cmd> ... (try 'jb run')")
	}

	switch os.Args[1] {

	case "run":
		// example: jb run -http=:8080 -mqtt=:1886
		//		run http and mqtt on non-std ports, bound to all interfaces
		// example: jb run -http=192.168.147.128:3333 -mqtt=192.168.147.128:1883
		// 		only run http and mqtt on 192.168.147.128 interface
		var httpAddr, mqttAddr string

		flagset := flag.NewFlagSet("runflags", flag.ContinueOnError)
		flagset.StringVar(&httpAddr, "http", "0.0.0.0:3333",
			"provide http server on <host><:port>")
		flagset.StringVar(&mqttAddr, "tcp", "0.0.0.0:1883",
			"*use* mqtt server on <host><:port>")
		flagset.Parse(os.Args[2:])

		if !strings.Contains(httpAddr, "://") {
			httpAddr = "http://" + httpAddr
		}
		hurl, err := url.Parse(httpAddr)
		check(err)

		// TODO the "-mqtt=..." flag will also be needed in other subcommands
		if !strings.Contains(mqttAddr, "://") {
			mqttAddr = "tcp://" + mqttAddr
		}
		murl, err := url.Parse(mqttAddr)
		check(err)

		startAllServers(hurl, murl)

	case "see":
		topics := "#"
		if len(os.Args) > 2 {
			topics = os.Args[2]
		}
		client = jeebus.NewClient(nil) // TODO -mqtt=... arg
		client.Register(topics, new(SeeService))
		<-client.Done

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
		client = jeebus.NewClient(nil) // TODO -mqtt=... arg

		//allow graceful closure from terminal etc.
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

		exit := make(chan bool)
		go func() {
			for sig := range sigchan {
				switch sig {
				case syscall.SIGINT:
					exit <- true
					log.Println("Exit via SIGINT")
				case syscall.SIGTERM:
					exit <- true
					log.Println("Exit via SIGTERM")
				}
			}
		}()

		//we task this as we get better coordination
		go serialConnect(dev, nbaud, tag, exit)

		<-client.Done
		// don't care about a graceful deregister since mqtt is gone
		log.Println("serial connection ends")

	case "tick":
		topic := "/admin/tick"
		if len(os.Args) > 2 {
			topic = os.Args[2]
		}
		client = jeebus.NewClient(nil) // TODO -mqtt=... arg
		go func() {
			ticker := time.NewTicker(time.Second)
			for tick := range ticker.C {
				client.Publish(topic, tick.String())
			}
		}()
		<-client.Done

	case "pub":
		if len(os.Args) < 3 {
			log.Fatalf("usage: jb pub <key> ?<jsonval>?")
		}
		var value string
		if len(os.Args) > 3 {
			value = os.Args[3]
		}
		client = jeebus.NewClient(nil) // TODO -mqtt=... arg
		client.Publish(os.Args[2], []byte(value))
		// TODO need to close gracefully, and not too soon!
		time.Sleep(10 * time.Millisecond)

	case "db":
		if len(os.Args) < 3 {
			log.Fatalf("usage: jb db <command> ...")
		}

		// o := &opt.Options{ ErrorIfMissing: true }
		dbf, err := leveldb.OpenFile("./storage", nil)
		check(err)
		db = dbf

		switch os.Args[2] {
		case "dump":
			switch len(os.Args) {
			case 3:
				dumpDatabase("", "")
			case 4:
				dumpDatabase(os.Args[3], "")
			case 5:
				dumpDatabase(os.Args[3], os.Args[4])
			}

		case "export":
			if len(os.Args) < 4 {
				log.Fatalf("usage: jb db export <prefix>")
			}
			exportJsonData(os.Args[3])

		case "import":
			if len(os.Args) < 4 {
				log.Fatalf("usage: jb db import <jsonfile>")
			}
			importJsonData(os.Args[3])

		case "compact":
			db.CompactRange(leveldb.Range{})

		default:
			log.Fatal("unknown db sub-command: jb db ", os.Args[2], " ...")
		}

	default:
		log.Fatal("unknown sub-command: jb ", os.Args[1], " ...")
	}
}

type SeeService struct{}

func (s *SeeService) Handle(m *jeebus.Message) {
	log.Println(m.T, string(m.P))
}

func exportJsonData(prefix string) {
	limit := prefix + "~" // FIXME see below, same as for dumpDatabase()
	entries := make(map[string]interface{})

	// get and print all the key/value pairs from the database
	iter := db.NewIterator(nil)
	iter.Seek([]byte(prefix))
	for iter.Valid() {
		key := iter.Key()[len(prefix):]
		if string(iter.Key()) > limit {
			break
		}
		var value interface{}
		err := json.Unmarshal(iter.Value(), &value)
		check(err)
		entries[string(key)] = value
		if !iter.Next() {
			break
		}
	}
	iter.Release()

	values := make(map[string]map[string]interface{})
	values[prefix] = entries

	s, e := json.MarshalIndent(values, "", "  ")
	check(e)
	fmt.Println(string(s))
}

func importJsonData(filename string) {
	data, err := ioutil.ReadFile(filename)
	check(err)

	var values map[string]map[string]*json.RawMessage
	err = json.Unmarshal(data, &values)
	check(err)

	for prefix, entries := range values {
		limit := prefix + "~" // FIXME see below, same as for dumpDatabase()
		var ndel, nadd int

		// get and print all the key/value pairs from the database
		iter := db.NewIterator(nil)
		iter.Seek([]byte(prefix))
		for iter.Valid() {
			key := string(iter.Key())
			if key > limit {
				break
			}
			err = db.Delete([]byte(key), nil)
			check(err)
			ndel++
			if !iter.Next() {
				break
			}
		}
		iter.Release()

		for k, v := range entries {
			err = db.Put([]byte(prefix+k), *v, nil)
			check(err)
			nadd++
		}

		fmt.Printf("%d deleted, %d added for prefix %q\n", ndel, nadd, prefix)
	}
}

func dumpDatabase(from, to string) {
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

func startAllServers(hurl, murl *url.URL) {
	var err error

	log.Println("opening database")
	db, err = leveldb.OpenFile("./storage", nil)
	check(err)

	log.Println("starting MQTT server on", murl)
	sock, err := net.Listen("tcp", murl.Host) // TODO tls!
	check(err)
	svr := mqtt.NewServer(sock)
	svr.Start()
	// <-svr.Done

	client = jeebus.NewClient(murl)
	client.Register("/#", &DatabaseService{})
	client.Register("io/+/+/+", new(LoggerService))
	client.Register("sv/lua/#", new(LuaDispatchService))
	client.Register("sv/rpc/#", new(RpcService))

	// persistent messages must be stored as JSON object
	client.Publish("/jb/info", map[string]interface{}{
    	"started": time.Now().Format(time.RFC822Z),
		"version": version,
		"webserver": hurl.String(),
    })

	// FIXME hook up the blinker script to handle incoming messages
	// FIXME broken due to recent change from rd/... to if/...
	// client.Publish("sv/lua/register", []byte("io/blinker/+/+"))

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for sig := range sigchan {
			switch sig {
			case syscall.SIGINT:
				log.Println("Exit via SIGINT")
				os.Exit(0)
			case syscall.SIGTERM:
				log.Println("Exit via SIGTERM")
				os.Exit(0)
				// case syscall.SIGHUP:
				// TODO this is where we can re-read config etc
			}
		}
	}()

	log.Println("starting web server on", hurl)
	http.Handle("/", http.FileServer(http.Dir("./app")))
	// TODO these extra access paths should probably not be hard-coded here
	fs := http.FileServer(http.Dir("./files"))
	http.Handle("/files/", http.StripPrefix("/files/", fs))
	lf := http.FileServer(http.Dir("./logger"))
	http.Handle("/logger/", http.StripPrefix("/logger/", lf))
	http.Handle("/ws", websocket.Handler(sockServer))
	log.Fatal(http.ListenAndServe(hurl.Host, nil)) // TODO https!
}

type DatabaseService struct{}

func (s *DatabaseService) Handle(m *jeebus.Message) {
	if len(m.P) > 0 {
		db.Put([]byte(m.T), m.P, nil)
		// TODO reconsider carefully whether to use timestamp inside payload
		millis := time.Now().UnixNano() / 1000000
		db.Put([]byte(fmt.Sprintf("hist/%s/%d", m.T, millis)), m.P, nil)
	} else {
		db.Delete([]byte(m.T), nil)
		// TODO decide what to do with deletions w.r.t. the historical data
		//  record the deletion? delete it as well? sweep and clean up later?
	}
	// send out websocket messages for all matching attached topics
	msg := make(map[string]*json.RawMessage)
	msg[m.T] = &m.P
	for k, v := range attached {
		if strings.HasPrefix(m.T, k) {
			for dest, _ := range v {
				client.Dispatch("ws/"+dest, msg) // direct dispatch, no MQTT
			}
		}
	}
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
	origin := tag + "/ip-" + ws.Request().RemoteAddr

	client.Register("ws/"+origin, &WebsocketService{ws})
	defer client.Unregister("ws/" + origin)

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
		// 	log.Println("shutdown requested from", origin)
		// 	os.Exit(0)
		case '"':
			// show incoming JSON strings on JB's stdout for debugging purposes
			var text string
			err := json.Unmarshal(msg, &text)
			check(err)
			log.Printf("%s (%s)", text, origin) // send to JB server's stdout
		case '[':
			// JSON array: either an MQTT publish request, or an RPC request
			var args []*json.RawMessage
			err := json.Unmarshal(msg, &args)
			check(err)
			var topic string
			if json.Unmarshal(*args[0], &topic) == nil {
				// it's an MQTT publish request
				log.Println("TOPIC", topic, args)
				if strings.HasPrefix(topic, "/") {
					client.Publish(topic, args[1])
				} else {
					log.Fatal("ws: topic must start with '/': ", topic)
				}
			} else {
				// it's an RPC request of the form (rpcId, req string, args...]
				msg = decodeRpcRequest(origin, msg)
				err = websocket.Message.Send(ws, string(msg))
				check(err)
			}
		default:
			// everything else (i.e. a JSON object) becomes an MQTT service req
			client.Publish("sv/"+origin, msg)
		}
	}
}

func decodeRpcRequest(name string, msg json.RawMessage) json.RawMessage {
	var any []interface{}
	err := json.Unmarshal(msg, &any)
	check(err)
	log.Printf("RPC %.100v (%s)", any, name)

	// process the RPC request, returns either a value or an error
	result, err := processRpcRequest(name, any[1].(string), any[2:])
	// convert errors to strings to send them through JSON
	var emsg interface{}
	if err != nil {
		emsg = err.Error()
		result = nil
	}
	reply := []interface{}{any[0], result, emsg}

	// shorten the reply log output to fit on one line
	logText := fmt.Sprintf("%v", reply)
	if len(logText) > 68 {
		logText = logText[:67] + "…"
	}
	log.Println(" ->", logText)

	out, err := json.Marshal(reply)
	check(err)
	return out
}

type RpcService struct{}

func (s *RpcService) Handle(m *jeebus.Message) {
	origin := strings.SplitN(m.T, "/", 3)[2]
	client.Publish("cb/"+origin, decodeRpcRequest(origin, m.P))
}

func processRpcRequest(name, cmd string, args []interface{}) (interface{}, error) {
	log.Printf("rpc cmd %s %v", cmd, args)
	switch cmd {

	case "echo":
		return args, nil

	case "db-keys":
		return dbKeys(args[0].(string)), nil

	case "db-get":
		v, e := db.Get([]byte(args[0].(string)), nil) // TODO yuck...
		return string(v), e

	case "lua":
		return luaRunWithArgs(args)

	case "attach":
		prefix := args[0].(string)
		if _, ok := attached[prefix]; !ok {
			attached[prefix] = make(map[string]int)
		}
		if _, ok := attached[prefix][name]; !ok {
			attached[prefix][name] = 0
		}
		attached[prefix][name]++
		log.Println("attached", prefix, name)

		to := prefix + "~" // TODO see notes about "~" elsewhere
		result := make(map[string]interface{})

		iter := db.NewIterator(nil)
		iter.Seek([]byte(prefix))
		for iter.Valid() {
			if string(iter.Key()) > to {
				break
			}
			var obj interface{}
			err := json.Unmarshal(iter.Value(), &obj) // TODO yuck, why decode?
			check(err)
			result[string(iter.Key())] = obj
			if !iter.Next() {
				break
			}
		}
		iter.Release()

		return result, nil

	case "detach":
		prefix := args[0].(string)
		if v, ok := attached[prefix]; ok {
			if _, ok := v[name]; ok {
				attached[prefix][name]--
				if attached[prefix][name] <= 0 {
					delete(attached[prefix], name)
					if len(attached[prefix]) == 0 {
						delete(attached, prefix)
					}
				}
			}
		}
		log.Println("detached", prefix, name)
		return nil, nil

	case "openfile":
		name := args[0].(string)
		// TODO this isn't safe if the filename uses a nasty path!
		return ioutil.ReadFile("files/" + name)

	case "savefile":
		name := args[0].(string)
		// TODO this isn't safe if the filename uses a nasty path!
		if len(args) > 1 {
			data := args[1].(string)
			log.Println("WRITE", "files/"+name)
			return nil, ioutil.WriteFile("files/"+name, []byte(data), 0666)
		} else {
			log.Println("REMOVE", "files/"+name)
			return nil, os.Remove("files/" + name)
		}
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
		if bytes.Compare(k, to) > 0 {
			break
		}
		i := bytes.IndexRune(k[skip:], '/') + skip
		if i < skip {
			i = len(k)
		}
		// fmt.Printf(" DK %d %d %d %s %s\n", skip, len(prev), i, prev, k)
		if !bytes.Equal(prev, k[skip:i]) {
			// need to make a copy of the key, since it's owned by iter
			prev = make([]byte, i-skip)
			copy(prev, k[skip:i])
			// fmt.Printf("ADD %s\n", prev)
			result = append(result, string(prev))
		}
		if !iter.Next() {
			break
		}
	}
	return result
}

type LoggerService struct{ fd *os.File }

// LOGGER_PREFIX is where log files get created. While this directory exists,
// the logger will store new files in it and append log items. Note that it is
// perfectly ok to create or remove this directory while the logger is running.
const LOGGER_PREFIX = "./logger/"

func (s *LoggerService) Handle(msg *jeebus.Message) {
	if !isPlainTextLine(msg.P) {
		return // filter input on if/... to only log simple plain text lines
	}
	
	split := strings.Split(msg.T, "/");
	port := split[2]
	// TODO accepting any value right now, but non-monotonic would be a problem
	//n, err := strconv.Atoi(split[3])
	n, err := strconv.ParseInt(split[3],10,64)	
	check(err)
	timestamp := time.Unix(0, int64(n) * 1000000)

	// automatic enabling/disabling of the logger, based on presence of dir
	_, err = os.Stat(LOGGER_PREFIX)
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
	datePath := dateFilename(timestamp)
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
	hour, min, sec := timestamp.Clock()
	line := fmt.Sprintf("L %02d:%02d:%02d.%03d %s %s\n",
		hour, min, sec, timestamp.Nanosecond()/1000000, port, msg.P)
	s.fd.WriteString(line)
}

func isPlainTextLine(input []byte) bool {
	if len(input) > 250 {
		return false // too long, limit is set (a bit arbitrarily) at 250 bytes
	}
	for _, b := range input {
		if b < 0x20 || b > 0x7E {
			return false // input has non-printable ASCII characters
		}
	}
	return true
}

func dateFilename(now time.Time) string {
	year, month, day := now.Date()
	path := fmt.Sprintf("%s%d", LOGGER_PREFIX, year)
	os.MkdirAll(path, os.ModePerm)
	// e.g. "./logger/2014/20140122.txt"
	return fmt.Sprintf("%s/%d.txt", path, (year*100+int(month))*100+day)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
