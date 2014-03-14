package network

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/workers"
)

func ExampleHTTPServer() {
	g := flow.NewGroup()
	g.Add("s", "HTTPServer")
	g.Set("s.Handlers", flow.Tag{"/blah/", "."})
	g.Set("s.Start", ":12345")
	g.Run()
	res, _ := http.Get("http://:12345/blah/http.go")
	body, _ := ioutil.ReadAll(res.Body)
	data, _ := ioutil.ReadFile("http.go")
	fmt.Println(string(body) == string(data))
	// Output:
	// Lost *url.URL: /blah/http.go
	// true
}
