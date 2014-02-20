package jeebus

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"code.google.com/p/go.net/websocket"
)

func init() {
	http.Handle("/ws", websocket.Handler(sockServer))
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
		var msg json.RawMessage
		err := websocket.JSON.Receive(ws, &msg)
		if err == io.EOF {
			break
		}
		Check(err)

		// figure out the structure of incoming JSON data to decide what to do
		switch msg[0] {
		// case 'n': // null
		// 	log.Println("shutdown requested from", origin)
		// 	os.Exit(0)
		case '"':
			// show incoming JSON strings on JB's stdout for debugging
			var text string
			err := json.Unmarshal(msg, &text)
			Check(err)
			log.Printf("%s (%s)", text, origin) // send to JB server's stdout
		case '[':
			// JSON array: either an MQTT publish request, or an RPC request
			var args []interface{}
			err := json.Unmarshal(msg, &args)
			Check(err)

			if topic, ok := args[0].(string); ok {
				// it's an MQTT publish request
				if strings.HasPrefix(topic, "/") {
					Publish(topic, args[1])
				} else {
					log.Fatal("ws: topic must start with '/': ", topic)
				}
			} else {
				// it's an RPC request of the form (rpcId, req string, args...]
				msg = decodeRpcRequest(origin, msg)
				err = websocket.Message.Send(ws, string(msg))
				Check(err)
			}
		default:
			// everything else (i.e. a JSON object) becomes an MQTT service req
			Publish("sv/"+origin, msg)
		}
	}
}

func decodeRpcRequest(name string, msg json.RawMessage) json.RawMessage {
	println(name, FromJson(msg))
	return []byte{}
}
