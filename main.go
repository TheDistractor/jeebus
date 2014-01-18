package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/syndtr/goleveldb/leveldb"
	// "github.com/jmhodges/levigo"
	"code.google.com/p/go.net/websocket"
	"github.com/aarzilli/golua/lua"
	"github.com/chimera/rs232"
	"github.com/jeffallen/mqtt"
	"github.com/stevedonovan/luar"
)

var (
	openConnections map[string]*websocket.Conn
	serialPort      *rs232.Port
)

func init() {
	openConnections = make(map[string]*websocket.Conn)
}

func main() {
	log.Println("opening database")
	openDatabase("./storage")

	log.Println("setting up Lua")
	setupLua()

	log.Println("starting MQTT server")
	go mqttServer()

	// passing serial port as first arg will override the default
	dev := "/dev/tty.usbserial-A40115A2"
	if len(os.Args) > 1 {
		dev = os.Args[1]
	}
	log.Println("opening serial port", dev)
	serialPort = serialConnect(dev)

	http.Handle("/", http.FileServer(http.Dir("public")))
	http.Handle("/ws", websocket.Handler(sockServer))
	log.Println("web server is listening on port 3333")
	log.Fatal(http.ListenAndServe("localhost:3333", nil))
}

func mqttServer() {
	port, err := net.Listen("tcp", ":1883")
	if err != nil {
		log.Fatal("listen: ", err)
	}
	svr := mqtt.NewServer(port)
	svr.Start()
	<-svr.Done
}

func openDatabase(dbname string) {
	db, err := leveldb.OpenFile(dbname, nil)
	// opts := levigo.NewOptions()
	// // opts.SetCache(levigo.NewLRUCache(1<<10))
	// opts.SetCreateIfMissing(true)
	// db, err := levigo.Open(dbname, opts)
	if err != nil {
		log.Fatal(err)
	}
	_ = db // ignore value for now
}

func serialConnect(dev string) *rs232.Port {
	// open the serial port
	options := rs232.Options{
		BitRate:  57600,
		DataBits: 8,
		StopBits: 1,
	}
	ser, err := rs232.Open(dev, options)
	if err != nil {
		log.Fatal(err)
	}

	// turn incoming data into a channel of text lines
	inputLines := make(chan string)

	go func() {
		scanner := bufio.NewScanner(ser)
		for scanner.Scan() {
			inputLines <- scanner.Text()
		}
	}()

	// process incoming data
	go func() {
		// flush all old data from the serial port
		log.Println("waiting for blinker to start")
		for line := range inputLines {
			if line == "[blinker]" {
				break
			}
			// TODO bail out if another sketch type is found
		}

		for line := range inputLines {
			log.Println(line)
			for _, conn := range openConnections {
				websocket.JSON.Send(conn, line)
			}
		}
	}()

	return ser
}

func sockServer(ws *websocket.Conn) {
	defer ws.Close()
	client := ws.Request().RemoteAddr
	openConnections[client] = ws
	log.Println("Client connected:", client)

	for {
		// accept arrays of arbitrary data types
		var any []interface{}
		err := websocket.JSON.Receive(ws, &any)
		if err != nil {
			log.Print(err)
			break
		}
		log.Println(any)

		// send as L<n><m> to the serial port
		cmd := fmt.Sprintf("L%.0f%.0f", any[0], any[1])
		serialPort.Write([]byte(cmd))
	}

	log.Println("Client disconnected:", client)
	delete(openConnections, client)
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

const testr = `
for i = 1,3 do
    Print(MSG,i)
end
`

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

	L.DoString(testr)
}
