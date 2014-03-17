package network

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/jcw/flow"
	_ "github.com/jcw/flow/gadgets"
)

func ExampleHTTPServer() {
	g := flow.NewCircuit()
	g.Add("s", "HTTPServer")
	g.Feed("s.Handlers", flow.Tag{"/blah/", "."})
	g.Feed("s.Start", ":12345")
	g.Run()
	res, _ := http.Get("http://:12345/blah/http.go")
	body, _ := ioutil.ReadAll(res.Body)
	data, _ := ioutil.ReadFile("http.go")
	fmt.Println(string(body) == string(data))
	// Output:
	// Lost *url.URL: /blah/http.go
	// true
}

func ExampleEnvVar() {
	os.Setenv("FOO", "bar!")

	g := flow.NewCircuit()
	g.Add("e", "EnvVar")
	g.Feed("e.In", "FOO")
	g.Feed("e.In", flow.Tag{"FOO", "def"})
	g.Feed("e.In", flow.Tag{"BLAH", "abc"})
	g.Run()
	// Output:
	// Lost string: bar!
	// Lost string: bar!
	// Lost string: abc
}
