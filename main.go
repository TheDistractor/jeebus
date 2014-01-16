package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"

	"code.google.com/p/go.net/websocket"
	"github.com/chimera/go-inside/rs232"
)

var connection *websocket.Conn

func main() {
	http.Handle("/", http.FileServer(http.Dir("public")))
	http.Handle("/ws", websocket.Handler(sockServer))

	// open the serial port
	options := rs232.Options{
		BitRate:  57600,
		DataBits: 8,
		StopBits: 1,
	}
	serial, err := rs232.Open("/dev/tty.usbserial-A40115A2", options)
	if err != nil {
		log.Fatal(err)
	}

	// turn incoming data into a channel of text lines
	input := make(chan string)
	go func() {
		scanner := bufio.NewScanner(serial)
		for scanner.Scan() {
			input <- scanner.Text()
		}
	}()

	// flush all old data from the serial port
	fmt.Println("waiting for blinker to start")
	for <-input != "[blinker]" {
		// TODO bail out if another sketch type is found
	}

	// process incoming data
	go func() {
		for {
			line := <-input
			fmt.Println(line)
			if connection != nil {
				websocket.JSON.Send(connection, line)
			}
		}
	}()

	println("listening on port 3333")
	log.Fatal(http.ListenAndServe("localhost:3333", nil))
}

func sockServer(ws *websocket.Conn) {
	defer ws.Close()
	connection = ws
	client := ws.Request().RemoteAddr
	log.Println("Client connected:", client)

	var msg struct {
		b string
		v int
	}

	for {
		var str string
		websocket.Message.Receive(ws, &str)
		fmt.Printf("str %v\n", str)

		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			log.Print(err)
			break
		}
		fmt.Printf("msg %v = %+v %s\n", client, msg, msg.b)
		websocket.JSON.Send(ws, "hi!")
	}
	log.Println("Client disconnected:", client)
	connection = nil
}
