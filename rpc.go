package jeebus

import (
	"errors"
	"strings"
)

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
		if cmd == "echo" {
			reply = args
		} else {
			err = errors.New("no such RPC command: " + cmd)
		}
	}

	replyFun(reply, err)
}
