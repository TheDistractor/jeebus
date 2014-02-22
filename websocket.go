package jeebus

import (
	"io"
	"log"
	"net/http"
	"strings"

	"code.google.com/p/go.net/websocket"
)

func init() {
	http.Handle("/ws", websocket.Handler(sockServer))
	
	http.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request){
		reload := ToJson(strings.HasSuffix(r.RequestURI, "?true"))
				
		// TODO: peeks into the messaging's services map, shouldn't be in here!
		for k, v := range services {
			if strings.HasPrefix(k, "ws/") {
				v.Handle("ws/<broadcast>", reload)
			}
		}
	})
}

type WebsocketService struct {
	ws *websocket.Conn
}

func (s *WebsocketService) Handle(topic string, payload []byte) {
	err := websocket.Message.Send(s.ws, string(payload))
	Check(err)
}

func sockServer(ws *websocket.Conn) {
	defer ws.Close()

	tag := ws.Request().Header.Get("Sec-Websocket-Protocol")
	origin := tag + "/ip-" + ws.Request().RemoteAddr

	Register("ws/"+origin, &WebsocketService{ws})
	defer Unregister("ws/" + origin)

	for {
		var msg interface{}
		err := websocket.JSON.Receive(ws, &msg)
		if err == io.EOF {
			break
		}
		Check(err)

		// examine the structure of the incoming request to decide what to do
		switch v := msg.(type) {

		case string:
			log.Printf("%s (%s)", v, origin) // send to JB server's stdout

		case []interface{}:
			// an array represents an RPC request or event (i.e. no reply)
			log.Printf("RPC %v (%s)", v, origin)
			if len(v) > 0 {
				if rpcId, ok := v[0].(float64); ok {
					// it's an RPC request with a return ID
					wsReplier := func(r interface{}, e error) {
						var emsg string
						if e != nil {
							emsg = e.Error()
						}
						reply := []interface{}{rpcId, r, emsg}
						err = websocket.Message.Send(ws, string(ToJson(reply)))
						Check(err)
					}
					// use a closure to encapsulate the reply handling
					ProcessRpc(origin, v[1:], wsReplier)
				} else {
					// it's an event without return ID, so no reply needed
					ProcessRpc(origin, v, dummyReplier)
				}
			} else {
				log.Println("empty [] request ignored")
			}

		default:
			// everything else becomes an MQTT service request
			Publish("sv/"+origin, msg)
		}
	}
}

func dummyReplier(r interface{}, e error) {
	Check(e)
}
