package jeebus_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/jcw/jeebus"
)

func TestFilesViaServer(t *testing.T) {
	response := serveOneRequest("GET", "/files/")
	expect(t, response.Code, http.StatusOK)
	refute(t, response.Body.Len(), 0)
}

func TestFileFetch(t *testing.T) {
	data := jeebus.Fetch("README.md")
	refute(t, len(data), 0)
}

func TestMissingFetch(t *testing.T) {
	data := jeebus.Fetch("blah")
	expect(t, len(data), 0)
}

func TestFileStore(t *testing.T) {
	err := jeebus.Store("hello", []byte("Hello, world!"))
	expect(t, err, nil)

	defer os.Remove(jeebus.Settings.FilesDir + "/hello")

	data, _ := ioutil.ReadFile(jeebus.Settings.FilesDir + "/hello")
	expect(t, string(data), "Hello, world!")
}

func TestFileDelete(t *testing.T) {
	err := jeebus.Store("foo", []byte("bar"))
	expect(t, err, nil)

	defer os.Remove(jeebus.Settings.FilesDir + "/foo")

	data, _ := ioutil.ReadFile(jeebus.Settings.FilesDir + "/foo")
	expect(t, string(data), "bar")

	err = jeebus.Store("foo", []byte{})
	expect(t, err, nil)

	_, err = ioutil.ReadFile(jeebus.Settings.FilesDir + "/foo")
	refute(t, err, nil)
}

func TestMissingDelete(t *testing.T) {
	err := jeebus.Store("bleep", []byte{})
	refute(t, err, nil)
}
