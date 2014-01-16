package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"

	"code.google.com/p/go.net/websocket"
	"github.com/chimera/go-inside/rs232"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("public")))
	mux.Handle("/engine.io/", websocket.Handler(wsHandler))

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
		}
	}()

	println("listening on port 3333")
	log.Fatal(http.ListenAndServe("localhost:3333", mux))
}

func wsHandler(ws *websocket.Conn) {
	io.Copy(ws, ws)
}
