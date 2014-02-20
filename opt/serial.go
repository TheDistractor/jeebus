package opt

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/chimera/rs232"
	"github.com/jcw/jeebus"
)

func init() {
	//
}

type SerialInterfaceService struct {
	io.ReadWriter
}

func (s *SerialInterfaceService) Handle(topic string, payload interface{}) {
	s.Write(payload.([]byte))
}

func serialConnect(port string, baudrate int, tag string) chan bool {
	// open the serial port in 8N1 mode
	opt := rs232.Options{BitRate: uint32(baudrate), DataBits: 8, StopBits: 1}
	dev, err := rs232.Open(port, opt)
	jeebus.Check(err)

	// flush old pending data
	avail, _ := dev.BytesAvailable()
	// from http://stackoverflow.com/questions/20320582
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
	var input, tag string
	var ok bool
	if tag, ok = matcher.(string); !ok {
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
	println(topic)

	jeebus.Register(topic, &SerialInterfaceService{dev})
	defer jeebus.Unregister(topic)

	// store the tag line for this device
	attachMsg := map[string]string{"text": input, "tag": tag}
	jeebus.Publish("/attach/"+name, attachMsg)
	defer jeebus.Publish("/detach/"+name, attachMsg)

	done := make(chan bool)

	// send the tag line (if present), then send out whatever comes in
	go func() {
		if input != "" {
			jeebus.Publish(topic, input)
		}
		for scanner.Scan() {
			jeebus.Publish(topic, scanner.Text())
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
	jeebus.Publish(topic, []byte(text)) // do not convert string to JSON
}
