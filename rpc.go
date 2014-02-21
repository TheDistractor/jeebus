package jeebus

import (
	"errors"
	"strings"
)

var rpcMap = make(map[string]func(string, []interface{}) interface{})

func init() {
	Define("echo", func(orig string, args []interface{}) interface{} {
		return args
	})
}

func Define(name string, cmdFun func(string, []interface{}) interface{}) {
	rpcMap[name] = cmdFun
}

func Undefine(name string) {
	delete(rpcMap, name)
}

// TODO: get rid of the orig arg, e.g. by switching to interfaces
func ProcessRpc(orig string, args []interface{}, replyFun func(r interface{}, e error)) {
	defer func() {
		if err, ok := recover().(error); ok && err != nil {
			replyFun(nil, err) // capture and report all panics
		}
	}()

	var reply interface{}
	var err error

	cmd := args[0].(string)
	args = args[1:]

	if f, ok := rpcMap[cmd]; ok {
		// TODO: add support for goroutines, i.e. replying later on
		reply = f(orig, args)
		// turn an error reply into a genuine error return
		if e, ok := reply.(error); ok {
			err = e
			reply = nil
		}
	} else if strings.HasPrefix(cmd, "/") {
		switch len(args) {
		case 0:
			reply = Fetch(cmd)
		case 1:
			Publish(cmd, args[0])
		default:
			err = errors.New("too many args")
		}
	} else {
		err = errors.New("unknown RPC command: " + cmd)
	}

	replyFun(reply, err)
}
