package main

import (
	"log"
	"os"
	"strconv"

	".." // the fully-specified path is "github.com/jcw/jeebus"
)

func main() {
	// TODO figure out how to use the "flag" package with sub-commands
	if len(os.Args) <= 1 {
		log.Fatal("usage: jb <cmd> ... (try 'jb run')")
	}

	switch os.Args[1] {
	case "run":
		jeebus.Server()

	case "see":
    topics := "#"
		if len(os.Args) > 2 {
			topics = os.Args[2]
		}
		seeCmd(topics)

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
		log.Fatal("unknown sub-command: jb ", os.Args[1], " ...")
	}
}

func seeCmd(topics string) {
	for m := range jeebus.ListenToServer(topics) {
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
