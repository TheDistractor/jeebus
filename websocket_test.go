package jeebus_test

import (
	"log"
	"net"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"code.google.com/p/go.net/websocket"
	"github.com/jcw/jeebus"
)

var (
	once       sync.Once
	serverAddr string
)

func startServer() {
	server := httptest.NewServer(nil)
	serverAddr = server.Listener.Addr().String()
	log.Print("Test WebSocket server listening on ", serverAddr)
}

func newClient(t *testing.T, proto string) net.Conn {
	client, err := net.Dial("tcp", serverAddr)
	expect(t, err, nil)

	myAddr := "http://" + client.LocalAddr().String()
	config, err := websocket.NewConfig("ws://"+serverAddr+"/ws", myAddr)
	config.Protocol = []string{proto}
	expect(t, err, nil)

	conn, err := websocket.NewClient(config, client)
	expect(t, err, nil)

	return conn
}

func TestWsSendMessage(t *testing.T) {
	once.Do(startServer)
	jeebus.StartMessaging()

	conn := newClient(t, "test")
	defer conn.Close()

	_, err := conn.Write([]byte(`"Hello, JeeBus"`)) // send over websocket
	expect(t, err, nil)

	// TODO: not checked, the message should appear on stdout
}

func TestWsPublishAndSave(t *testing.T) {
	once.Do(startServer)
	jeebus.StartMessaging()

	conn := newClient(t, "test")
	defer conn.Close()

	spy := make(SpyService, 1)
	jeebus.Register("/foo", &spy)
	defer jeebus.Unregister("/foo")
	defer jeebus.Put("/foo", nil)

	_, err := conn.Write([]byte(`["/foo", "bar"]`)) // send over websocket
	expect(t, err, nil)

	reply := <-spy
	expect(t, reply.a, "/foo")                 // the published message came back
	expect(t, jeebus.FromJson(reply.b), "bar") // the published message came back

	any := jeebus.Get("/foo")
	expect(t, any, "bar") // the data ended up in the database
}

func TestWsServiceRequest(t *testing.T) {
	once.Do(startServer)
	jeebus.StartMessaging()

	conn := newClient(t, "test")
	defer conn.Close()
	myAddr := conn.LocalAddr().String()

	spy := make(SpyService, 1)
	jeebus.Register("sv/test/#", &spy)
	defer jeebus.Unregister("sv/test/#")

	_, err := conn.Write([]byte(`{"foo": "bar"}`)) // send over websocket
	expect(t, err, nil)

	reply := <-spy
	expect(t, reply.a, strings.Replace(myAddr, "http://", "sv/test/ip-", 1))
	expect(t, string(reply.b), `{"foo": "bar"}`)
}

func TestWsIsRegistered(t *testing.T) {
	// var actual_msg = make([]byte, 512)
	// n, err := conn.Read(actual_msg)
	// expect(t, err, nil)
	// expect(t, n, 0)
	// expect(t, jeebus.IsRegistered("ws/#"), true)
}
