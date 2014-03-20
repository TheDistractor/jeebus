package network

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/golang/glog"
	"github.com/jcw/flow"
)

func init() {
	flow.Registry["HTTPServer"] = func() flow.Circuitry { return &HTTPServer{} }

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
	Start    flow.Input
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
	port := (<-w.Start).(string)
	if _, err := strconv.Atoi(port); err == nil {
		port = ":" + port // convert "1234" -> ":1234"
	}
	go func() {
		// will stay running until an error is returned or the app ends
		defer flow.DontPanic()
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
		w.Out.Send(msg)
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
