package network

import (
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
	flow.Registry["EnvVar"] = func() flow.Circuitry { return &EnvVar{} }
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
	mux.HandleFunc("/reload", reloadHandler)
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

// broadcast a reload request to each attached websocket client
func reloadHandler(w http.ResponseWriter, req *http.Request) {
	reload := req.URL.RawQuery == "true"
	for _, ws := range wsClients {
		websocket.JSON.Send(ws, reload)
	}
	w.Write([]byte{})
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

// Lookup an environment variable, with optional default.
type EnvVar struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
}

// Start lookup up environment variables.
func (g *EnvVar) Run() {
	for m := range g.In {
		switch v := m.(type) {
		case string:
			m = os.Getenv(v)
		case flow.Tag:
			if s := os.Getenv(v.Tag); s != "" {
				m = s
			} else {
				m = v.Msg
			}
		}
		g.Out.Send(m)
	}
}
