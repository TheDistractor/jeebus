package jeebus_test

import (
	"testing"

	"github.com/jcw/jeebus"
)

var reply string

func mockReply(t *testing.T) func(r interface{}, e error) {
	return func(r interface{}, e error) {
		if e == nil {
			reply = string(jeebus.ToJson(r))
		} else {
			reply = "Error: " + e.Error()
		}
	}
}

func wrap(args ...interface{}) []interface{} {
	return args
}

func TestEchoRpc(t *testing.T) {
	jeebus.ProcessRpc(wrap("echo", 1, 2, 3, 4.5), mockReply(t))
	expect(t, reply, "[1,2,3,4.5]")
}

func TestFetchMissing(t *testing.T) {
	jeebus.ProcessRpc(wrap("/blah"), mockReply(t))
	expect(t, reply, "")
}

func TestTooManyArgs(t *testing.T) {
	jeebus.ProcessRpc(wrap("/blah", 1, 2), mockReply(t))
	expect(t, reply, "Error: too many args")
}
