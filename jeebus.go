package jeebus

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var Version = "0.3.0"

var nonPrintables = regexp.MustCompile("[^[:print:]]")

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

func DisplayMax(any interface{}, maxLen int) string {
	s := fmt.Sprintf("%v", any)
	s = nonPrintables.ReplaceAllLiteralString(s, ".")
	if len(s) > maxLen {
		s = s[:maxLen-1] + "…"
	}
	return s
}
