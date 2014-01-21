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
		port := ":3333"
		if len(os.Args) > 2 {
			port = os.Args[2]
		}
		jeebus.Server(port)

	case "see":
		topics := "#"
		if len(os.Args) > 2 {
			topics = os.Args[2]
		}
		for m := range jeebus.ListenToServer(topics) {
			log.Println(m.T, string(m.P.([]byte)), m.R)
		}

	case "serial":
		if len(os.Args) <= 2 {
			log.Fatal("usage: jb serial <dev> ?baud? ?tag?")
		}
		dev, baud, tag := os.Args[2], "57600", ""
		if len(os.Args) > 3 {
			baud = os.Args[3]
		}
		if len(os.Args) > 4 {
			tag = os.Args[4]
		}
		nbaud, err := strconv.Atoi(baud)
		if err != nil {
			log.Fatal(err)
		}
		serialCmd(dev, nbaud, tag)

	default:
		log.Fatal("unknown sub-command: jb ", os.Args[1], " ...")
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
