package jeebus

import (
	"errors"
	"log"
	"strings"
)

func ProcessRpc(args []interface{}, replyFun func(r interface{}, e error)) {
	var cmd string
	if len(args) > 0 {
		cmd = args[0].(string)
		args = args[1:]
	}
	log.Println("cmd", cmd, "args", args)
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
	}

	if cmd == "echo" {
		reply = args
	}

	replyFun(reply, err)
}
