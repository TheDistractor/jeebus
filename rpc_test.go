package jeebus_test

import (
	"testing"

	"github.com/jcw/jeebus"
)

var reply interface{}

func mockReply(t *testing.T) func(r interface{}, e error) {
	reply = ""
	return func(r interface{}, e error) {
		if e == nil {
			reply = r
		} else {
			reply = "ERR: " + e.Error()
		}
	}
}

func wrapArgs(args ...interface{}) []interface{} {
	return args
}

func TestEchoRpc(t *testing.T) {
	jeebus.ProcessRpc(wrapArgs("echo", 1, "2", 3, 4.5), mockReply(t))
	expect(t, string(jeebus.ToJson(reply)), `[1,"2",3,4.5]`)
}

func TestFetchMissingRpc(t *testing.T) {
	jeebus.ProcessRpc(wrapArgs("/blah"), mockReply(t))
	expect(t, string(jeebus.ToJson(reply)), "")
}

func TestTooManyArgsRpc(t *testing.T) {
	jeebus.ProcessRpc(wrapArgs("/blah", 1, 2), mockReply(t))
	expect(t, reply, "ERR: too many args")
}

func TestBadRpc(t *testing.T) {
	jeebus.ProcessRpc(wrapArgs(1, 2), mockReply(t))
	expect(t, reply, "ERR: interface conversion: interface is int, not string")
}

func TestEmptyRpc(t *testing.T) {
	jeebus.ProcessRpc(wrapArgs(), mockReply(t))
	expect(t, reply, "ERR: runtime error: index out of range")
}

func TestUnknownRpc(t *testing.T) {
	jeebus.ProcessRpc(wrapArgs("blah"), mockReply(t))
	expect(t, reply, "ERR: unknown RPC command: blah")
}

func TestDefineRpc(t *testing.T) {
	jeebus.ProcessRpc(wrapArgs("twice", 11), mockReply(t))
	expect(t, reply, "ERR: unknown RPC command: twice")

	jeebus.Define("twice", func(args []interface{}) interface{} {
		return 2 * args[0].(int)
	})

	jeebus.ProcessRpc(wrapArgs("twice", 22), mockReply(t))
	expect(t, reply, 44)

	jeebus.Undefine("twice")

	jeebus.ProcessRpc(wrapArgs("twice", 33), mockReply(t))
	expect(t, reply, "ERR: unknown RPC command: twice")
}

func TestDbGetRpc(t *testing.T) {
	jeebus.ProcessRpc(wrapArgs("db-get", "blah"), mockReply(t))
	expect(t, reply, nil)
}

func TestDbKeysRpc(t *testing.T) {
	jeebus.ProcessRpc(wrapArgs("db-keys", "blah"), mockReply(t))
	expect(t, string(jeebus.ToJson(reply)), "null") // JSON unmarshaling quirk
	expect(t, len(reply.([]string)), 0) 			// this is a better check
}
