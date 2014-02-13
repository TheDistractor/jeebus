// Separate serial to abstract rs232 to get better x-platform support.
package main

import (
	"bufio"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chimera/rs232"
	"github.com/jcw/jeebus"
)

type SerialInterfaceService struct {
	*rs232.Port
}

func (s *SerialInterfaceService) Handle(m *jeebus.Message) {
	s.Write([]byte(m.Get("text")))
}

func serialConnect(port string, baudrate int, tag string, done chan bool) {
	// open the serial port in 8N1 mode
	serial, err := rs232.Open(port, rs232.Options{
		BitRate: uint32(baudrate), DataBits: 8, StopBits: 1,
	})
	check(err)

	scanner := bufio.NewScanner(serial)

	var input string;

	// flush all old data from the serial port while looking for a tag

	log.Println("waiting for serial")
	timeout := time.Now().Add(10*time.Second) // TODO: turn into --timeout=10
	if tag == "" {
		for scanner.Scan() {
			if time.Now().After(timeout) {
				log.Println("Serial Timeout obtaining tag.")
				client.Done <- true
				return  // no need to detach as it was never attached
			}
			input = scanner.Text()
			if strings.HasPrefix(input, "[") &&
				strings.Contains(input, "]") {
				tag = input[1:strings.IndexAny(input, ".]")]
				break
			}

		}
	}

	dev := strings.TrimPrefix(port, "/dev/")
	dev = strings.Replace(dev, "tty.usbserial-", "usb-", 1)
	name := "io/" + tag + "/" + dev
	log.Println("serial ready:", name)

	client.Register(name, &SerialInterfaceService{serial})
	defer client.Unregister(name)

	// store the tag line for this device
	attachMsg := map[string]string{"text": input, "tag": tag}
	client.Publish("/attach/"+dev, attachMsg)
	defer client.Publish("/detach/"+dev, attachMsg)

	// send the tag line (if present), then send out whatever comes in
	if input != "" {
		publishWithTimeStamp(name, input)
	}
	for scanner.Scan() {
		select {
		case <-done:
			client.Publish("/detach/"+dev, attachMsg)
			serial.Close()
		default:
			publishWithTimeStamp(name, scanner.Text())
		}
	}
	
	log.Println("Serial Disconnect!!")
	<-time.After(2 * time.Second) // allow things to naturally close
	client.Done <- true

}

func publishWithTimeStamp(tag, text string) {
	now := time.Now().UTC().UnixNano() / 1000000
	topic := fmt.Sprintf("%s/%d", tag, now)
	client.Publish(topic, []byte(text)) // do not convert string to JSON
}
