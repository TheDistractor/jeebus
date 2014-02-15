// The JeeBus server, with messaging, data storage, and a web server.
package main

import (
	"encoding/json"
	"fmt"
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
	"github.com/codegangsta/cli"
	"github.com/jcw/jeebus"
	"github.com/jeffallen/mqtt"
	"github.com/syndtr/goleveldb/leveldb"
	// "github.com/syndtr/goleveldb/leveldb/opt"
)

const version = "0.2-beta"

var (
	db     *leveldb.DB
	client *jeebus.Client
)

func init() {
	log.SetFlags(log.Ltime)
}

func main() {
	app := cli.NewApp()
	app.Name = "jeebus"
	app.Version = version
	app.Usage = "Messaging and data storage infrastructure for low-end systems."

	app.Flags = []cli.Flag{
		cli.StringFlag{"mqtt, m", "1883",
			"MQTT server port (external if specified as <host:port>)"},
	}

	app.Commands = []cli.Command{
		{
			Name:   "run",
			Usage:  "Launch as server with HTTP, WebSockets, and MQTT",
			Action: runCommand,
			Flags: []cli.Flag{
				cli.StringFlag{"port, p", "3333",
					"HTTP server port (limited to interface if <iface:port>)"},
			},
		},
		{
			Name:   "see",
			Usage:  "Listen to server messages and display them",
			Action: seeCommand,
		},
		{
			Name:   "serial",
			Usage:  "Connect a serial port to the server",
			Action: serialCommand,
		},
		{
			Name:   "tick",
			Usage:  "Publish periodic ticks (useful for debugging)",
			Action: tickCommand,
		},
		{
			Name:   "pub",
			Usage:  "Publish a single message to the server",
			Action: pubCommand,
		},
		{
			Name:   "dump",
			Usage:  "Dump the contents of the database",
			Action: dumpCommand,
		},
		{
			Name:   "export",
			Usage:  "Export a range from the database as JSON",
			Action: exportCommand,
		},
		{
			Name:   "import",
			Usage:  "Import a range from JSON into the database",
			Action: importCommand,
		},
		{
			Name:   "compact",
			Usage:  "Perform a database compaction",
			Action: compactCommand,
		},
	}

	app.Run(os.Args)
}

func runCommand(c *cli.Context) {
	// example: jb run -http=:8080 -mqtt=:1886
	//		run http and mqtt on non-std ports, bound to all interfaces
	// example: jb run -http=192.168.147.128:3333 -mqtt=192.168.147.128:1883
	// 		only run http and mqtt on 192.168.147.128 interface
	mqttAddr := c.GlobalString("mqtt")
	httpAddr := c.String("port")
	// TODO: the "-mqtt=..." flag will also be needed in other subcommands
	startAllServers(asUrl(httpAddr, "http"), asUrl(mqttAddr, "tcp"))
}

func asUrl(addr, proto string) *url.URL {
	if _, err := strconv.Atoi(addr); err == nil {
		addr = ":" + addr
	}
	if !strings.Contains(addr, "://") {
		addr = proto + "://" + addr
	}
	u, err := url.Parse(addr)
	check(err)
	return u
}

func seeCommand(c *cli.Context) {
	topics := c.Args().First()
	if topics == "" {
		topics = "#"
	}
	client = jeebus.NewClient(nil) // TODO: -mqtt=... arg
	client.Register(topics, new(SeeService))
	<-client.Done
}

func serialCommand(c *cli.Context) {
	dev, baud, tag := c.Args().Get(0), c.Args().Get(1), c.Args().Get(2)
	if dev == "" {
		log.Fatalf("usage: jb serial <dev> ?baud? ?tag?")
	}
	if baud == "" {
		baud = "57600"
	}

	nbaud, err := strconv.Atoi(baud)
	check(err)
	log.Printf("opening serial port %s @ %d baud", dev, nbaud)
	client = jeebus.NewClient(nil) // TODO: -mqtt=... arg

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
}

func tickCommand(c *cli.Context) {
	topic := c.Args().First()
	if topic == "" {
		topic = "/admin/tick"
	}
	client = jeebus.NewClient(nil) // TODO: -mqtt=... arg
	go func() {
		ticker := time.NewTicker(time.Second)
		for tick := range ticker.C {
			client.Publish(topic, tick.String())
		}
	}()
	<-client.Done
}

func pubCommand(c *cli.Context) {
	if len(c.Args()) < 2 {
		log.Fatalf("usage: jb pub <topic> ?<jsonval>?")
	}
	topic, value := c.Args().Get(0), c.Args().Get(1)
	client = jeebus.NewClient(nil) // TODO: -mqtt=... arg
	client.Publish(topic, []byte(value))
	// TODO: need to close gracefully, and not too soon!
	time.Sleep(10 * time.Millisecond)
}

func dumpCommand(c *cli.Context) {
	from, to := c.Args().Get(0), c.Args().Get(1)
	if to == "" {
		to = from + "~" // FIXME this assumes all key chars are less than "~"
	}

	openDatabase()
	// get and print all the key/value pairs from the database
	iter := db.NewIterator(nil)
	iter.Seek([]byte(from))
	for iter.Valid() {
		if string(iter.Key()) > to {
			break
		}
		fmt.Printf("%s = %s\n", iter.Key(), iter.Value())
		if !iter.Next() {
			break
		}
	}
	iter.Release()
}

func exportCommand(c *cli.Context) {
	if len(c.Args()) < 1 {
		log.Fatalf("usage: jb export <prefix>")
	}
	openDatabase()
	exportJsonData(c.Args().First())
}

func importCommand(c *cli.Context) {
	if len(c.Args()) < 1 {
		log.Fatalf("usage: jb import <jsonfile>")
	}
	openDatabase()
	importJsonData(c.Args().First())
}

func compactCommand(c *cli.Context) {
	openDatabase()
	db.CompactRange(leveldb.Range{})
}

func openDatabase() {
	// o := &opt.Options{ ErrorIfMissing: true }
	var err error
	db, err = leveldb.OpenFile("./storage", nil)
	check(err)
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

func startAllServers(hurl, murl *url.URL) {
	openDatabase()

	log.Println("starting MQTT server on", murl)
	sock, err := net.Listen("tcp", murl.Host) // TODO: tls!
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
		"started":   time.Now().Format(time.RFC822Z),
		"version":   version,
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
				// TODO: this is where we can re-read config etc
			}
		}
	}()

	log.Println("starting web server on", hurl)
	http.Handle("/", http.FileServer(http.Dir("./app")))
	// TODO: these extra access paths should probably not be hard-coded here
	fs := http.FileServer(http.Dir("./files"))
	http.Handle("/files/", http.StripPrefix("/files/", fs))
	lf := http.FileServer(http.Dir("./logger"))
	http.Handle("/logger/", http.StripPrefix("/logger/", lf))
	http.Handle("/ws", websocket.Handler(sockServer))
	log.Fatal(http.ListenAndServe(hurl.Host, nil)) // TODO: https!
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
