package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"

	"code.google.com/p/go.net/websocket"
	"github.com/chimera/rs232"
	"github.com/jeffallen/mqtt"
)

var (
	openConnections map[string]*websocket.Conn
	serialPort      *rs232.Port
)

func init() {
	openConnections = make(map[string]*websocket.Conn)
}

func main() {
	serialConnect("/dev/tty.usbserial-A40115A2")

	go startMqttServer()

	println("listening on port 3333")
	http.Handle("/", http.FileServer(http.Dir("public")))
	http.Handle("/ws", websocket.Handler(sockServer))
	log.Fatal(http.ListenAndServe("localhost:3333", nil))
}

func startMqttServer() {
	fmt.Println("starting MQTT server")
	l, err := net.Listen("tcp", ":1883")
	if err != nil {
		log.Fatal("listen: ", err)
	}
	svr := mqtt.NewServer(l)
	svr.Start()
	<-svr.Done
}

func serialConnect(dev string) {
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
	serialPort = ser

	// turn incoming data into a channel of text lines
	inputLines := make(chan string)

	go func() {
		scanner := bufio.NewScanner(serialPort)
		for scanner.Scan() {
			inputLines <- scanner.Text()
		}
	}()

	// process incoming data
	go func() {
		// flush all old data from the serial port
		fmt.Println("waiting for blinker to start")
		for line := range inputLines {
			if line == "[blinker]" {
				break
			}
			// TODO bail out if another sketch type is found
		}

		for line := range inputLines {
			fmt.Println(line)
			for _, conn := range openConnections {
				websocket.JSON.Send(conn, line)
			}
		}
	}()
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
		fmt.Println(any)

		// send as L<n><m> to the serial port
		cmd := fmt.Sprintf("L%.0f%.0f", any[0], any[1])
		serialPort.Write([]byte(cmd))
	}

	log.Println("Client disconnected:", client)
	delete(openConnections, client)
}
