package main

import (
	"log"
	"os"
	"strconv"

	".." // the fully-specified path is "github.com/jcw/jeebus"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "see":
			seeCmd()
		case "serial":
			if len(os.Args) < 4 {
				log.Fatal("usage: jb serial <dev> <baud> ?tag?")
			}
			dev, sbaud, tag := os.Args[2], os.Args[3], ""
			if len(os.Args) > 4 {
				tag = os.Args[4]
			}
			baud, err := strconv.Atoi(sbaud)
			if err != nil {
				log.Fatal(err)
			}
			serialCmd(dev, baud, tag)
		default:
			jeebus.Server()
		}
	} else {
		jeebus.Server()
	}
}

func seeCmd() {
	for m := range jeebus.ListenToServer("#") {
		topic := m.T
		message := m.P
		retain := ""
		if m.R {
			retain = "(retain)"
		}
		log.Println(topic, retain, string(message.([]byte)))
	}
}

func serialCmd(dev string, baud int, tag string) {
	feed := jeebus.ListenToServer("if/serial")

	log.Println("opening serial port", dev)
	serial := jeebus.SerialConnect(dev, baud, tag)

	for m := range feed {
		log.Printf("Ser: %s", m.P.([]byte))
		serial.Write(m.P.([]byte))
	}
}
