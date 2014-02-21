package jeebus_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/jcw/jeebus"
)

func TestFilesViaServer(t *testing.T) {
	response := serveOneRequest("GET", "/files/")
	expect(t, response.Code, http.StatusOK)
	refute(t, response.Body.Len(), 0)
}

func TestFileFetch(t *testing.T) {
	fn := jeebus.Settings.FilesDir + "/ping"
	defer os.Remove(fn)

	err := ioutil.WriteFile(fn, []byte("pong"), 0644)
	expect(t, err, nil)

	data := jeebus.Fetch("ping")
	expect(t, string(data), "pong")
}

func TestMissingFetch(t *testing.T) {
	data := jeebus.Fetch("blah")
	expect(t, len(data), 0)
}

func TestFileStore(t *testing.T) {
	fn := jeebus.Settings.FilesDir + "/hello"
	defer os.Remove(fn)

	err := jeebus.Store("hello", []byte("Hello, world!"))
	expect(t, err, nil)

	data, _ := ioutil.ReadFile(fn)
	expect(t, string(data), "Hello, world!")
}

func TestFileDelete(t *testing.T) {
	fn := jeebus.Settings.FilesDir + "/foo"
	defer os.Remove(fn)

	err := jeebus.Store("foo", []byte("bar"))
	expect(t, err, nil)

	data, _ := ioutil.ReadFile(fn)
	expect(t, string(data), "bar")

	err = jeebus.Store("foo", []byte{})
	expect(t, err, nil)

	_, err = ioutil.ReadFile(fn)
	refute(t, err, nil)
}

func TestMissingDelete(t *testing.T) {
	err := jeebus.Store("bleep", []byte{})
	refute(t, err, nil)
}

func TestFileList(t *testing.T) {
	fn := jeebus.Settings.FilesDir + "/foo"
	defer os.Remove(fn)

	err := ioutil.WriteFile(fn, []byte("bar"), 0644)
	expect(t, err, nil)

	files := jeebus.FileList(".", false)
	refute(t, len(files), 0)
	expect(t, contains(files, "foo"), true)
}

func TestMissingDirList(t *testing.T) {
	files := jeebus.FileList("blah", false)
	expect(t, len(files), 0)

	dirs := jeebus.FileList("blah", true)
	expect(t, len(dirs), 0)
}

func TestSafePaths(t *testing.T) {
	var tests = []struct {
		path, clean string
		safe        bool
	}{
		{"", ".", false},
		{".", ".", false},
		{"..", "..", false},
		{"a", "a", true},
		{"a/..", ".", false},
		{"a/../b", "b", true},
		{"a/.", "a", true},
		{"a/", "a", true},
		{"/", "/", false},
	}

	for _, x := range tests {
		c := path.Clean(x.path)
		s := jeebus.PathIsSafe(x.path)
		if c != x.clean || s != x.safe {
			t.Errorf("path '%s' clean '%s' safe %v - got '%s' (%v)",
				x.path, x.clean, x.safe, c, s)
		}
	}
}

func TestSubdirStore(t *testing.T) {
	fn := jeebus.Settings.FilesDir + "/hey"
	defer os.RemoveAll(fn)

	err := jeebus.Store("hey/hello", []byte("Howdy!"))
	expect(t, err, nil)

	data, _ := ioutil.ReadFile(fn + "/hello")
	expect(t, string(data), "Howdy!")

	expect(t, contains(jeebus.FileList(".", false), "hey"), false)
	expect(t, contains(jeebus.FileList(".", true), "hey"), true)
	expect(t, contains(jeebus.FileList("hey", false), "hello"), true)
	expect(t, contains(jeebus.FileList("hey", true), "hello"), false)
}

func TestSubdirDelete(t *testing.T) {
	fn := jeebus.Settings.FilesDir + "/hey"
	defer os.RemoveAll(fn)

	err := jeebus.Store("hey/hello", []byte("Yes!"))
	expect(t, err, nil)

	data, _ := ioutil.ReadFile(fn + "/hello")
	expect(t, string(data), "Yes!")

	err = jeebus.Store("hey/hi", []byte("No?"))
	expect(t, err, nil)

	data, _ = ioutil.ReadFile(fn + "/hi")
	expect(t, string(data), "No?")

	// subdir exists, with both files in it
	expect(t, contains(jeebus.FileList(".", true), "hey"), true)
	expect(t, contains(jeebus.FileList("hey", false), "hello"), true)
	expect(t, contains(jeebus.FileList("hey", false), "hi"), true)

	err = jeebus.Store("hey/hello", []byte{})
	expect(t, err, nil)

	// subdir exists, with only one file in it
	expect(t, contains(jeebus.FileList(".", true), "hey"), true)
	expect(t, contains(jeebus.FileList("hey", false), "hello"), false)
	expect(t, contains(jeebus.FileList("hey", false), "hi"), true)

	err = jeebus.Store("hey/hi", []byte{})
	expect(t, err, nil)

	// subdir should no longer exist
	expect(t, contains(jeebus.FileList(".", true), "hey"), false)
}
