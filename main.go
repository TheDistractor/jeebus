package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"

	"code.google.com/p/go.net/websocket"
	"github.com/chimera/rs232"
)

var (
	openConnections map[string]*websocket.Conn
	serialPort      *rs232.Port
)

func init() {
	openConnections = make(map[string]*websocket.Conn)
}

func main() {
	// open the serial port
	options := rs232.Options{
		BitRate:  57600,
		DataBits: 8,
		StopBits: 1,
	}
	ser, err := rs232.Open("/dev/tty.usbserial-A40115A2", options)
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

	http.Handle("/", http.FileServer(http.Dir("public")))
	http.Handle("/ws", websocket.Handler(sockServer))

	println("listening on port 3333")
	log.Fatal(http.ListenAndServe("localhost:3333", nil))
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
