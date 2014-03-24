package network

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jcw/flow"
	_ "github.com/jcw/flow/gadgets"
)

func ExampleHTTPServer() {
	g := flow.NewCircuit()
	g.Add("s", "HTTPServer")
	g.Feed("s.Handlers", flow.Tag{"/blah/", "."})
	g.Feed("s.Port", ":12345")
	g.Run()
	// time.Sleep(50 * time.Millisecond)
	res, _ := http.Get("http://:12345/blah/http.go")
	body, _ := ioutil.ReadAll(res.Body)
	data, _ := ioutil.ReadFile("http.go")
	fmt.Println(string(body) == string(data))
	// Output:
	// Lost *url.URL: /blah/http.go
	// true
}
