package jeebus

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/codegangsta/cli"
)

var (
	Version       = "0.3.0"
	nonPrintables = regexp.MustCompile("[^[:print:]]")
	app           *cli.App
)

func NewApp(name, version string) *cli.App {
	app = cli.NewApp()
	app.Name = name
	app.Version = version

	return app
}

func NewCommand(cmd *cli.Command) {
	if app == nil {
		NewApp("jeebus", Version)
	}

	app.Commands = append(app.Commands, *cmd)
}

func Run() {
	DefineToolCommands()
	app.Run(os.Args)
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
