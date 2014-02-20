package jeebus_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/jcw/jeebus"
)

func TestLoggerIsRegistered(t *testing.T) {
	ok := jeebus.IsRegistered("io/+/+/+")
	expect(t, ok, true)
}

func TestLoggerViaServer(t *testing.T) {
	response := serveOneRequest("GET", "/logger/")
	expect(t, response.Code, http.StatusOK)
	refute(t, response.Body.Len(), 0)
}

func TestLogger(t *testing.T) {
	jeebus.StartMessaging()

	spy := newSpyService()
	jeebus.Register("io/test/#", &spy)
	defer jeebus.Unregister("io/test/#")

	// 1234567890 ms = Jan 15, 1970 07:56:07.890 UTC
	jeebus.Publish("io/test/dummy/1234567890", []byte("ping"))
	defer os.RemoveAll("logger/1970")

	reply := <-spy
	expect(t, reply.a, "io/test/dummy/1234567890")
	expect(t, string(reply.b.([]byte)), "ping")

	_, err := ioutil.ReadFile("logger/1970/19700115.txt")
	expect(t, err, nil)

	// TODO: check contents of file (it's currently bad: TWO entries?)
	// expected: L 07:56:07.890 dummy ping
}
