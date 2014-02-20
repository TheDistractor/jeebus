package jeebus

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/chimera/rs232"
)

func init() {
	Register("/serial/+/#", &SerialConnectService{})
}

type SerialConnectService struct{}

func (s *SerialConnectService) Handle(topic string, payload []byte) {
	args := strings.SplitN(topic, "/", 3)
	switch args[2] {
	case "connect":
		serialConnect(args[3], 57600, "") // TODO: real config settings
	}
}

type SerialInterfaceService struct {
	io.ReadWriter
}

func (s *SerialInterfaceService) Handle(topic string, payload []byte) {
	s.Write(payload)
}

func serialConnect(port string, baudrate int, tag string) chan bool {
	// open the serial port in 8N1 mode
	opt := rs232.Options{BitRate: uint32(baudrate), DataBits: 8, StopBits: 1}
	dev, err := rs232.Open(port, opt)
	Check(err)

	// flush old pending data
	avail, _ := dev.BytesAvailable()
	io.CopyN(ioutil.Discard, dev, int64(avail))

	if tag == "" {
		return SerialHandler(port, dev, &JeeTagMatcher{})
	} else {
		return SerialHandler(port, dev, tag)
	}
}

type SerialMatcher interface {
	Match(s string, dev io.Writer) string
}

type JeeTagMatcher struct{}

func (m *JeeTagMatcher) Match(s string, dev io.Writer) (tag string) {
	// TODO: add support for a timeout
	if strings.HasPrefix(s, "[") &&
		strings.Contains(s, "]") {
		tag = s[1:strings.IndexAny(s, ".]")]
	}
	return
}

// type KnownTagMatcher struct{
// 	tag string
// }
//
// func (m *KnownTagMatcher) Match(s string, dev io.Writer) string {
// 	return m.tag
// }

func SerialHandler(port string, dev io.ReadWriter, matcher interface{}) chan bool {
	scanner := bufio.NewScanner(dev)

	// use matcher as tag without scanning, if it was passed in as string
	var input string
	tag, ok := matcher.(string)
	if !ok {
		m := matcher.(SerialMatcher)
		log.Println("waiting for serial")
		// flush all old data from the serial port while looking for a tag
		for tag == "" && scanner.Scan() {
			input = scanner.Text()
			tag = m.Match(input, dev)
		}
		log.Println("serial tag found:", tag)
	}

	name := shortSerialName(port)
	topic := "io/" + tag + "/" + name

	Register(topic, &SerialInterfaceService{dev})
	defer Unregister(topic)

	// store the tag line for this device
	attachMsg := map[string]string{"text": input, "tag": tag}
	Publish("/attach/"+name, attachMsg)
	defer Publish("/detach/"+name, attachMsg)

	done := make(chan bool)

	// send the tag line (if present), then send out whatever comes in
	go func() {
		if input != "" {
			publishWithTimeStamp(topic, input)
		}
		for scanner.Scan() {
			publishWithTimeStamp(topic, scanner.Text())
		}
		close(done)
	}()

	return done
}

func shortSerialName(s string) string {
	dev := strings.TrimPrefix(s, "/dev/")
	return strings.Replace(dev, "tty.usbserial-", "usb-", 1)
}

func publishWithTimeStamp(tag, text string) {
	now := time.Now().UTC().UnixNano() / 1000000
	topic := fmt.Sprintf("%s/%d", tag, now)
	Publish(topic, []byte(text)) // do not convert string to JSON
}
