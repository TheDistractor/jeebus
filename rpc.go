package jeebus

import (
	"errors"
	"strings"
)

var rpcs = make(map[string]func(string,[]interface{}) (interface{}, error))

func init() {
	Define("echo", func(cmd string, args []interface{}) (interface{}, error) {
		return args, nil
	})
}

func Define(name string, cmdFun func(string,[]interface{}) (interface{}, error)) {
	rpcs[name] = cmdFun
}

func Undefine(name string) {
	delete(rpcs, name)
}

func ProcessRpc(args []interface{}, replyFun func(r interface{}, e error)) {
	defer func() {
		if err, ok := recover().(error); ok && err != nil {
			replyFun(0, err)
		}
	}()

	if len(args) == 0 {
		replyFun(nil, errors.New("empty RPC command ignored"))
		return
	}

	cmd := args[0].(string)
	args = args[1:]

	var reply interface{}
	var err error

	if strings.HasPrefix(cmd, "/") {
		switch len(args) {
		case 0:
			reply = Fetch(cmd)
		case 1:
			Publish(cmd, args[0])
		default:
			err = errors.New("too many args")
		}
	} else {
		if f, ok := rpcs[cmd]; ok {
			reply, err = f(cmd, args)
		} else {
			err = errors.New("no such RPC command: " + cmd)
		}
	}

	replyFun(reply, err)
}
