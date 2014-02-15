// Standard services, available for use in the JeeBus server.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/jcw/jeebus"
)

var attached = make(map[string]map[string]int) //prefix -> tag -> refcount

type SeeService struct{}

func (s *SeeService) Handle(m *jeebus.Message) {
	log.Println(m.T, string(m.P))
}

type DatabaseService struct{}

func (s *DatabaseService) Handle(m *jeebus.Message) {
	if len(m.P) > 0 {
		db.Put([]byte(m.T), m.P, nil)
		// TODO: reconsider carefully whether to use timestamp inside payload
		millis := time.Now().UnixNano() / 1000000
		db.Put([]byte(fmt.Sprintf("hist/%s/%d", m.T, millis)), m.P, nil)
	} else {
		db.Delete([]byte(m.T), nil)
		// TODO: decide what to do with deletions w.r.t. the historical data
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
	ws *websocket.Conn // TODO: can't this struct nesting be avoided, somehow?
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
		v, e := db.Get([]byte(args[0].(string)), nil) // TODO: yuck...
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

		to := prefix + "~" // TODO: see notes about "~" elsewhere
		result := make(map[string]interface{})

		iter := db.NewIterator(nil)
		iter.Seek([]byte(prefix))
		for iter.Valid() {
			if string(iter.Key()) > to {
				break
			}
			var obj interface{}
			err := json.Unmarshal(iter.Value(), &obj) // TODO: yuck, why decode?
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
		// TODO: this isn't safe if the filename uses a nasty path!
		return ioutil.ReadFile("files/" + name)

	case "savefile":
		name := args[0].(string)
		// TODO: this isn't safe if the filename uses a nasty path!
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
	// TODO: decide whether this key logic is the most useful & least confusing
	// TODO: should use skips and reverse iterators once the db gets larger!
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

	split := strings.Split(msg.T, "/")
	port := split[2]
	// TODO: accepting any value right now, but non-monotonic would be a problem
	n, err := strconv.ParseInt(split[3], 10, 64)
	check(err)
	timestamp := time.Unix(0, int64(n)*1000000)

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
