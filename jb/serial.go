// Separate serial to abstract rs232 to get better x-platform support.
package main

import (
	"bufio"
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

	var input struct {
		Text string `json:"text"`
		Time int64  `json:"time"`
	}

	// flush all old data from the serial port while looking for a tag

	log.Println("waiting for serial")
	timeout := time.Now().Add(10*time.Second) // TODO turn into --timeout=10
	if tag == "" {
		for scanner.Scan() {
			if time.Now().After(timeout) {
				log.Println("Serial Timeout obtaining tag.")
				client.Done <- true
				return  // no need to detach as it was never attached
			}
			input.Time = jeebus.TimeStamp(time.Now())
			input.Text = scanner.Text()
			if strings.HasPrefix(input.Text, "[") &&
				strings.Contains(input.Text, "]") {
				tag = input.Text[1:strings.IndexAny(input.Text, ".]")]
				break
			}

		}
	}

	dev := strings.TrimPrefix(port, "/dev/")
	dev = strings.Replace(dev, "tty.usbserial-", "usb-", 1)
	name := tag + "/" + dev
	log.Println("serial ready:", name)

	client.Register("if/"+name, &SerialInterfaceService{serial})
	defer client.Unregister("if/" + name)

	// store the tag line for this device
	attachMsg := map[string]string{"text": input.Text, "tag": tag}
	client.Publish("/attach/"+dev, attachMsg)
	defer client.Publish("/detach/"+dev, attachMsg)

	// send the tag line (if present), then send out whatever comes in
	if input.Text != "" {
		client.Publish("rd/"+name, &input)
	}
	for scanner.Scan() {
		select {
		case <-done:
			// don't defer since mqtt may be gone
			client.Publish("/detach/"+dev, attachMsg)
			// client.Publish("rm/unregister", map[string]string{"Tag":
			//	tag, "Id": dev, "Endpoint": resp.Endpoint})
			serial.Close()

		default:
			input.Time = jeebus.TimeStamp(time.Now())
			input.Text = scanner.Text()
			client.Publish("rd/"+name, &input) //see new logger topic
			// client.Publish(resp.Endpoint+"/"+
			//	jeebus.TimeStampToString(input.Time), &input)

		}
	}
	<-time.After(2 * time.Second) // allow things to naturally close
	log.Println("Serial Disconnect!!")
	client.Done <- true

}
