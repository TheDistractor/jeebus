package network

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/golang/glog"
	"github.com/jcw/flow"
)

func init() {
	flow.Registry["HTTPServer"] = func() flow.Circuitry { return &HTTPServer{} }
	flow.Registry["RpcHandler"] = func() flow.Circuitry { return &RpcHandler{} }

	// websockets without Sec-Websocket-Protocol are connected in loopback mode
	flow.Registry["WebSocket-default"] = flow.Registry["Pipe"]

	// use a special channel to pick up JSON "ipc" messages from stdin
	// this is currently used to broadcast reload triggers to all websockets
	go func() {
		// TODO: turn into a gadget, so that this can also be used with MQTT
		for m := range ipcFromNodeJs() {
			// FIXME: yuck, the JSON parsing is immediately re-encoded below!
			// can't send a []byte, since this sends as binary msg iso JSON
			var any interface{}
			if err := json.Unmarshal(m, &any); err == nil {
				for _, ws := range wsClients {
					websocket.JSON.Send(ws, any)
				}
			}
		}
	}()
}

// hack alert! special code to pick up node.js live reload triggers
// listens to stdin for a special "null" request and other JSON messages
// the initial "null" triggers sending this process's PID back to node.js
// nothing bad happens if stdin is closed or no data ever come in
func ipcFromNodeJs() chan []byte {
	ipc := make(chan []byte)
	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		if scanner.Scan() {
			if scanner.Text() == "null" {
				// yes, this is *writing* to stdin, i.e. used as IPC mechanism!
				os.Stdin.Write([]byte(fmt.Sprintf("%d\n", os.Getpid())))
			} else {
				ipc <- scanner.Bytes()
			}
		}
		for scanner.Scan() {
			ipc <- scanner.Bytes()
		}
	}()
	return ipc
}

var wsClients = map[string]*websocket.Conn{}

// HTTPServer is a .Feed( which sets up an HTTP server.
type HTTPServer struct {
	flow.Gadget
	Handlers flow.Input
	Port     flow.Input
	Out      flow.Output
}

type flowHandler struct {
	h http.Handler
	s *HTTPServer
}

func (fh *flowHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fh.s.Out.Send(req.URL)
	fh.h.ServeHTTP(w, req)
}

// Set up the handlers, then start the server and start processing requests.
func (w *HTTPServer) Run() {
	mux := http.NewServeMux() // don't use default to allow multiple instances
	for m := range w.Handlers {
		tag := m.(flow.Tag)
		switch v := tag.Msg.(type) {
		case string:
			h := createHandler(tag.Tag, v)
			mux.Handle(tag.Tag, &flowHandler{h, w})
		case http.Handler:
			mux.Handle(tag.Tag, &flowHandler{v, w})
		}
	}

	port := getInputOrConfig(w.Port, "HTTP_PORT")
	go func() {
		// will stay running until an error is returned or the app ends
		defer flow.DontPanic()
		glog.Infoln("http started on", port)
		err := http.ListenAndServe(port, mux)
		glog.Fatal(err)
	}()
	// TODO: this is a hack to make sure the server is ready
	// better would be to interlock the goroutine with the listener being ready
	time.Sleep(10 * time.Millisecond)
}

func createHandler(tag, s string) http.Handler {
	// TODO: hook gadget in as HTTP handler
	// if _, ok := flow.Registry[s]; ok {
	// 	return http.Handler(reqHandler)
	// }
	if s == "<websocket>" {
		return websocket.Handler(wsHandler)
	}
	if !strings.ContainsAny(s, "./") {
		glog.Fatalln("cannot create handler for:", s)
	}
	h := http.FileServer(http.Dir(s))
	if s != "/" {
		h = http.StripPrefix(tag, h)
	}
	if tag != "/" {
		return h
	}
	// special-cased to return main page unless the URL has an extension
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if path.Ext(r.URL.Path) == "" {
			r.URL.Path = "/"
		}
		h.ServeHTTP(w, r)
	})
}

func wsHandler(ws *websocket.Conn) {
	defer flow.DontPanic()
	defer ws.Close()

	hdr := ws.Request().Header

	// keep track of connected clients for reload broadcasting
	id := hdr.Get("Sec-Websocket-Key")
	wsClients[id] = ws
	defer delete(wsClients, id)

	// the protocol name is used as tag to locate the proper circuit
	tag := hdr.Get("Sec-Websocket-Protocol")
	if tag == "" {
		tag = "default"
	}

	g := flow.NewCircuit()
	g.AddCircuitry("head", &wsHead{ws: ws})
	g.Add("ws", "WebSocket-"+tag)
	g.AddCircuitry("tail", &wsTail{ws: ws})
	g.Connect("head.Out", "ws.In", 0)
	g.Connect("ws.Out", "tail.In", 0)
	g.Run()
}

type wsHead struct {
	flow.Gadget
	Out flow.Output

	ws *websocket.Conn
}

func (w *wsHead) Run() {
	for {
		var msg interface{}
		err := websocket.JSON.Receive(w.ws, &msg)
		if err == io.EOF {
			break
		}
		flow.Check(err)
		if s, ok := msg.(string); ok {
			id := w.ws.Request().Header.Get("Sec-Websocket-Key")
			fmt.Println("msg <"+id[:4]+">:", s)
		} else {
			w.Out.Send(msg)
		}
	}
}

type wsTail struct {
	flow.Gadget
	In flow.Input

	ws *websocket.Conn
}

func (w *wsTail) Run() {
	for m := range w.In {
		err := websocket.JSON.Send(w.ws, m)
		flow.Check(err)
	}
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
