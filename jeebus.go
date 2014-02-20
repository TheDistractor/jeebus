package jeebus

import (
	"encoding/json"
)

var Version = "0.3.0"

func Run() {
	OpenDatabase()
	StartMessaging()
	RunHttpServer()
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func ToJson(value interface{}) json.RawMessage {
	switch v := value.(type) {
	case []byte:
		return v
	case json.RawMessage:
		return v
	default:
		data, err := json.Marshal(&value)
		Check(err)
		return data
	}
}

func FromJson(value json.RawMessage) interface{} {
	var any interface{}
	err := json.Unmarshal(value, &any)
	Check(err)
	return any
}
