package jeebus_test

import (
	"io/ioutil"
	"os"
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

func TestFetchRpc(t *testing.T) {
	fn := jeebus.Settings.FilesDir + "/ping"
	defer os.Remove(fn)

	err := ioutil.WriteFile(fn, []byte("pong"), 0644)
	expect(t, err, nil)

	jeebus.ProcessRpc(wrapArgs("fetch", "ping"), mockReply(t))
	expect(t, string(reply.([]byte)), "pong")
}

func TestStoreRpc(t *testing.T) {
	fn := jeebus.Settings.FilesDir + "/foo"
	defer os.Remove(fn)

	jeebus.ProcessRpc(wrapArgs("store", "foo", []byte("bar")), mockReply(t))
	expect(t, reply, nil)
}

func TestRemoveMissingRpc(t *testing.T) {
	jeebus.ProcessRpc(wrapArgs("store", "foo", []byte{}), mockReply(t))
	expect(t, reply, "ERR: remove ./files/foo: no such file or directory")
}

func TestFileListRpc(t *testing.T) {
	fn := jeebus.Settings.FilesDir + "/ping"
	defer os.Remove(fn)

	err := ioutil.WriteFile(fn, []byte("pong"), 0644)
	expect(t, err, nil)

	jeebus.ProcessRpc(wrapArgs("file-list", ".", false), mockReply(t))
	expect(t, contains(reply.([]string), "ping"), true)
}
