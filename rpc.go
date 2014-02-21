package jeebus

import (
	"errors"
	"strings"
)

var rpcMap = make(map[string]func([]interface{}) interface{})

func init() {
	Define("echo", func(args []interface{}) interface{} {
		return args
	})
	Define("db-get", func(args []interface{}) interface{} {
		return Get(args[0].(string))
	})
	Define("db-keys", func(args []interface{}) interface{} {
		return Keys(args[0].(string))
	})
}

func Define(name string, cmdFun func([]interface{}) interface{}) {
	rpcMap[name] = cmdFun
}

func Undefine(name string) {
	delete(rpcMap, name)
}

func ProcessRpc(args []interface{}, replyFun func(r interface{}, e error)) {
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
		reply = f(args)
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
