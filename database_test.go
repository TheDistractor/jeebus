package jeebus_test

import (
	"os"
	"testing"

	"github.com/jcw/jeebus"
)

func TestOpenDatabase(t *testing.T) {
	err := os.RemoveAll("./db")
	expect(t, err, nil)

	err = jeebus.OpenDatabase()
	expect(t, err, nil)

	_, err = os.Open("./db")
	expect(t, err, nil)
}

func TestNonExistent(t *testing.T) {
	any := jeebus.Get("blah")
	expect(t, any, nil)
}

func TestStoreAndDelete(t *testing.T) {
	jeebus.Put("foo", "bar")
	any := jeebus.Get("foo")
	expect(t, any, "bar")
	jeebus.Put("foo", nil)
	any = jeebus.Get("foo")
	expect(t, any, nil)
}

func TestStoreViaPublish(t *testing.T) {
	onceStartMessaging(t)

	spy := newSpyService()
	jeebus.Register("/blah", &spy)
	defer jeebus.Unregister("/blah")

	jeebus.Publish("/blah", "bleep")
	reply := <-spy
	expect(t, reply.a, "/blah")
	expect(t, reply.b, "bleep")

	any := jeebus.Get("/blah")
	expect(t, any, "bleep")
	jeebus.Put("/blah", nil)
	any = jeebus.Get("/blah")
	expect(t, any, nil)
}

func TestNoStoreViaPublish(t *testing.T) {
	onceStartMessaging(t)

	spy := newSpyService()
	jeebus.Register("blah", &spy)
	defer jeebus.Unregister("blah")

	jeebus.Publish("blah", "bleep")
	reply := <-spy
	expect(t, reply.a, "blah")
	expect(t, reply.b, "bleep")

	any := jeebus.Get("blah")
	expect(t, any, nil)
}

func TestNoKeys(t *testing.T) {
	keys := jeebus.Keys("blah")
	expect(t, len(keys), 0)
}

func TestOneKey(t *testing.T) {
	jeebus.Put("a/b", 1)
	defer jeebus.Put("a/b", nil)

	keys := jeebus.Keys("a/")
	expect(t, len(keys), 1)
	expect(t, keys[0], "b")
}

func TestManyKeys(t *testing.T) {
	jeebus.Put("a/b1", 1)
	jeebus.Put("a/b2", 2)
	jeebus.Put("a/b3", 3)
	defer jeebus.Put("a/b1", nil)
	defer jeebus.Put("a/b2", nil)
	defer jeebus.Put("a/b3", nil)

	keys := jeebus.Keys("a/")
	expect(t, len(keys), 3)
	expect(t, keys[0], "b1")
	expect(t, keys[1], "b2")
	expect(t, keys[2], "b3")
}

func TestNestedKeys(t *testing.T) {
	jeebus.Put("a/b1", 1)
	jeebus.Put("a/b2/c1", 2)
	jeebus.Put("a/b2/c2", 3)
	jeebus.Put("a/b3", 4)
	defer jeebus.Put("a/b1", nil)
	defer jeebus.Put("a/b2/c1", nil)
	defer jeebus.Put("a/b2/c2", nil)
	defer jeebus.Put("a/b3", nil)

	keys := jeebus.Keys("a/")
	expect(t, len(keys), 3)
	expect(t, keys[0], "b1")
	expect(t, keys[1], "b2")
	expect(t, keys[2], "b3")

	keys = jeebus.Keys("a/b2/")
	expect(t, len(keys), 2)
	expect(t, keys[0], "c1")
	expect(t, keys[1], "c2")
}

func TestEmptyPartialKey(t *testing.T) {
	jeebus.Put("a/b", 1)
	defer jeebus.Put("a/b", nil)

	keys := jeebus.Keys("a")
	expect(t, len(keys), 1)
	expect(t, keys[0], "") // only up to the next "/"!
}

func TestPartialKey(t *testing.T) {
	jeebus.Put("ab", 1)
	defer jeebus.Put("ab", nil)

	keys := jeebus.Keys("a")
	expect(t, len(keys), 1)
	expect(t, keys[0], "b")
}

func TestManyPartialKeys(t *testing.T) {
	jeebus.Put("a1", 1)
	jeebus.Put("a2", 2)
	jeebus.Put("a3", 3)
	defer jeebus.Put("a1", nil)
	defer jeebus.Put("a2", nil)
	defer jeebus.Put("a3", nil)

	keys := jeebus.Keys("a")
	expect(t, len(keys), 3)
	expect(t, keys[0], "1")
	expect(t, keys[1], "2")
	expect(t, keys[2], "3")
}

func TestNestedPartialKey(t *testing.T) {
	jeebus.Put("a1/b1", 1)
	jeebus.Put("a1/b2", 2)
	jeebus.Put("a2/b3", 3)
	defer jeebus.Put("a1/b1", nil)
	defer jeebus.Put("a1/b2", nil)
	defer jeebus.Put("a2/b3", nil)

	keys := jeebus.Keys("a")
	expect(t, len(keys), 2)
	expect(t, keys[0], "1")
	expect(t, keys[1], "2")
}