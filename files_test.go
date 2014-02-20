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
	data := jeebus.Fetch("README.md")
	refute(t, len(data), 0)
}

func TestMissingFetch(t *testing.T) {
	data := jeebus.Fetch("blah")
	expect(t, len(data), 0)
}

func TestFileStore(t *testing.T) {
	defer os.Remove(jeebus.Settings.FilesDir + "/hello")

	err := jeebus.Store("hello", []byte("Hello, world!"))
	expect(t, err, nil)

	data, _ := ioutil.ReadFile(jeebus.Settings.FilesDir + "/hello")
	expect(t, string(data), "Hello, world!")
}

func TestFileDelete(t *testing.T) {
	defer os.Remove(jeebus.Settings.FilesDir + "/foo")

	err := jeebus.Store("foo", []byte("bar"))
	expect(t, err, nil)

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

func contains(list []string, value string) bool {
	for _, x := range list {
		if x == value {
			return true
		}
	}
	return false
}

func TestFileList(t *testing.T) {
	files := jeebus.FileList(".", false)
	refute(t, len(files), 0)
	expect(t, contains(files, "README.md"), true)
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
	defer os.RemoveAll(jeebus.Settings.FilesDir + "/hey")

	err := jeebus.Store("hey/hello", []byte("Howdy!"))
	expect(t, err, nil)

	data, _ := ioutil.ReadFile(jeebus.Settings.FilesDir + "/hey/hello")
	expect(t, string(data), "Howdy!")

	expect(t, contains(jeebus.FileList(".", false), "hey"), false)
	expect(t, contains(jeebus.FileList(".", true), "hey"), true)
	expect(t, contains(jeebus.FileList("hey", false), "hello"), true)
	expect(t, contains(jeebus.FileList("hey", true), "hello"), false)
}

func TestSubdirDelete(t *testing.T) {
	defer os.RemoveAll(jeebus.Settings.FilesDir + "/hey")

	err := jeebus.Store("hey/hello", []byte("Yes!"))
	expect(t, err, nil)

	data, _ := ioutil.ReadFile(jeebus.Settings.FilesDir + "/hey/hello")
	expect(t, string(data), "Yes!")

	err = jeebus.Store("hey/hi", []byte("No?"))
	expect(t, err, nil)

	data, _ = ioutil.ReadFile(jeebus.Settings.FilesDir + "/hey/hi")
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
