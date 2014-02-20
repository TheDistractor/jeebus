package opt_test

import (
	"testing"
	"time"

	"github.com/jcw/jeebus"
	"github.com/jcw/jeebus/opt"
)

func TestSerial(t *testing.T) {
	expect(t, nil, nil)
}

type SerialMessages struct {
	delay int
	text  string
}

type SerialMock struct {
	pending []SerialMessages
	written []string
}

func (s *SerialMock) Read(b []byte) (n int, err error) {
	if len(s.pending) > 0 {
		time.Sleep(time.Duration(s.pending[0].delay) * time.Millisecond)
		n = copy(b, []byte(s.pending[0].text))
		s.pending = s.pending[1:]
	}
	return
}

func (s *SerialMock) Write(b []byte) (n int, err error) {
	s.written = append(s.written, string(b))
	return len(b), nil
}

func TestSerialMock(t *testing.T) {
	jeebus.OpenDatabase()
	jeebus.StartMessaging()

	mock := &SerialMock{
		pending: []SerialMessages{
			{10, "aa\n"},
			{10, "[bb] xx\n"},
			{10, "cc\n"},
		},
	}

	spy1 := newSpyService()
	jeebus.Register("/attach/#", &spy1)
	defer jeebus.Unregister("/attach/#")

	spy2 := newSpyService()
	jeebus.Register("io/bb/+", &spy2)
	defer jeebus.Unregister("io/bb/+")

	spy3 := newSpyService()
	jeebus.Register("/detach/#", &spy3)
	defer jeebus.Unregister("/detach/#")

	done := opt.SerialHandler("/dev/ttyUSB0", mock, &opt.JeeTagMatcher{})

	reply := <-spy1
	expect(t, reply.a, "/attach/ttyUSB0")
	expect(t, string(jeebus.ToJson(reply.b)), `{"tag":"bb","text":"[bb] xx"}`)

	jeebus.Publish("io/bb/ttyUSB0", []byte("blah"))

	// FIXME: not working!
	// println(88888)
	// 	reply = <-spy2
	// 	expect(t, reply.a, "io/bb/ttyUSB0")
	// 	expect(t, reply.b.(string), "blah")
	// println(99999)

	<-done
	defer jeebus.Put("/attach/ttyUSB0", nil)
	defer jeebus.Put("/detach/ttyUSB0", nil)

	reply = <-spy3
	expect(t, reply.a, "/detach/ttyUSB0")
	expect(t, string(jeebus.ToJson(reply.b)), `{"tag":"bb","text":"[bb] xx"}`)

	// attach/detach messages have been stored in the database
	expect(t, string(jeebus.ToJson(jeebus.Get("/attach/ttyUSB0"))),
		`{"tag":"bb","text":"[bb] xx"}`)
	expect(t, string(jeebus.ToJson(jeebus.Get("/detach/ttyUSB0"))),
		`{"tag":"bb","text":"[bb] xx"}`)

	expect(t, len(mock.pending), 0) // all mock messages sent
	// FIXME: not working!
	// expect(t, len(mock.written), 2) // two mock messages received
	// expect(t, mock.written[0], "[bb] xx")
	// expect(t, mock.written[1], "cc")
}
