package network

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/jcw/flow"
)

func init() {
	flow.Registry["RpcHandler"] = func() flow.Circuitry { return new(RpcHandler) }
}

// RpcHandler turns incoming messages into RPC calls and send out the results.
type RpcHandler struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
}

// Start waiting for RPC requests.
func (g *RpcHandler) Run() {
	for m := range g.In {
		if rpc, ok := m.([]interface{}); ok && len(rpc) >= 2 {
			if cmd, ok := rpc[0].(string); ok {
				m = g.handleRpcRequest(cmd, int(rpc[1].(float64)), rpc[2:])
			}
		}
		g.Out.Send(m)
	}
}

func (g *RpcHandler) handleRpcRequest(cmd string, seq int, args []interface{}) (reply []interface{}) {
	if cmd == "echo" {
		return []interface{}{seq, "", args}
	}

	defer func() {
		errMsg := ""
		switch v := recover().(type) {
		case nil:
			// no error
		case string:
			errMsg = v
		case error:
			errMsg = v.Error()
		default:
			errMsg = fmt.Sprintf("%T: %v", v, v)
		}
		if errMsg != "" {
			glog.Warningln("rpc-error", cmd, args, errMsg)
			reply = []interface{}{seq, errMsg}
		}
	}()

	// if there's registered circuit for cmd, set it up and return as a stream
	fmt.Println("RPC:", cmd, args)
	if _, ok := flow.Registry[cmd]; ok && len(args) == 1 {
		c := flow.NewCircuit()
		c.Add("x", cmd)
		c.AddCircuitry("y", &streamRpcResults{seqNum: seq, replies: g})
		c.Connect("x.Out", "y.In", 0)
		for k, v := range args[0].(map[string]interface{}) {
			c.Feed("x."+k, tryToConvertToTag(v))
		}
		go func() {
			defer flow.DontPanic()
			c.Run()
			g.Out.Send([]interface{}{seq, false}) // end streaming
		}()
		return []interface{}{seq, true} // start streaming
	}

	panic(cmd + "?")
}

func tryToConvertToTag(v interface{}) interface{} {
	if t, ok := v.(map[string]interface{}); ok && len(t) == 2 {
		if tag, ok := t["Tag"]; ok {
			if msg, ok := t["Msg"]; ok {
				v = flow.Tag{tag.(string), msg}
			}
		}
	}
	return v
}

type streamRpcResults struct {
	flow.Gadget
	In flow.Input

	seqNum  int
	replies *RpcHandler
}

func (g *streamRpcResults) Run() {
	for m := range g.In {
		g.replies.Out.Send([]interface{}{g.seqNum, "Out", m})
	}
}
