package main

import (
	"log"
	"os"
	"strconv"

	".."
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "see":
			jeebus.SeeCmd()
		case "serial":
			if len(os.Args) < 4 {
				log.Fatal("usage: jeebus serial <dev> <baud> ?tag?")
			}
			dev, sbaud, tag := os.Args[2], os.Args[3], ""
			if len(os.Args) > 4 {
				tag = os.Args[4]
			}
			baud, err := strconv.Atoi(sbaud)
			if err != nil {
				log.Fatal(err)
			}
			jeebus.SerialCmd(dev, baud, tag)
		default:
			jeebus.Server()
		}
	} else {
		jeebus.Server()
	}
}
