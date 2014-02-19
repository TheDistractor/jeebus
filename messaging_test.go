package jeebus_test

import (
	"net"
	"testing"

	"github.com/jcw/jeebus"
)

func TestStartMessaging(t *testing.T) {
	onceStartMessaging(t)
}

func TestConnectToServer(t *testing.T) {
	conn, err := net.Dial("tcp", ":1883")
	expect(t, err, nil)

	err = conn.Close()
	expect(t, err, nil)
}

type HelloService chan interface{}

func (s *HelloService) Handle(topic string, payload interface{}) {
	*s <- payload
}

func TestPublish(t *testing.T) {
	hello := HelloService(make(chan interface{}, 1))
	jeebus.Register("hello", &hello)
	defer jeebus.Unregister("hello")
	jeebus.Publish("hello", "world")
	reply := <-hello
	expect(t, reply, "world")
}

/* TODO: doesn't work, gives "failed to submit message" errors after 25 calls
func BenchmarkPublish(b *testing.B) {
	hello := HelloService(make(chan interface{}, 1))
	jeebus.Register("hello", &hello)
	defer jeebus.Unregister("hello")

	for i := 0; i < b.N; i++ {
		jeebus.Publish("hello", "world")
		<-hello
	}
}
*/

// this was adapted from jeffallen/mqtt's test code, with thanks
func TestMatch(t *testing.T) {
	var tests = []struct {
		pattern, topic string
		expected       bool
	}{
		{"finance/stock/ibm/#", "finance/stock", false},
		{"finance/stock/ibm/#", "", false},
		{"finance/stock/ibm/#", "finance/stock/ibm", true},
		{"finance/stock/ibm/#", "", false},
		{"#", "anything", true},
		{"#", "anything/no/matter/how/deep", true},
		{"", "", true},
		{"+/#", "one", true},
		{"+/#", "", false}, // TODO: incorrect, probably?
		{"finance/stock/+/close", "finance/stock", false},
		{"finance/stock/+/close", "finance/stock/ibm", false},
		{"finance/stock/+/close", "finance/stock/ibm/close", true},
		{"finance/stock/+/close", "finance/stock/ibm/open", false},
		{"+/+/+", "", false},
		{"+/+/+", "a/b", false},
		{"+/+/+", "a/b/c", true},
		{"+/+/+", "a/b/c/d", false},
	}

	for _, x := range tests {
		got := jeebus.MatchTopic(x.pattern, x.topic)
		if got != x.expected {
			t.Error("Fail:", x.pattern, x.topic, "got", got)
		}
	}
}
