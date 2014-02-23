package jeebus_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/jcw/jeebus"
)

func TestVersion(t *testing.T) {
	expect(t, jeebus.Version, "0.3.0")
}

func TestCheck(t *testing.T) {
	jeebus.Check(nil)
	expect(t, failCheck(), "boom?")
}

func failCheck() (ret string) {
	// capture the error and add a question mark to the result
	defer func() { ret = recover().(error).Error() + "?" }()
	jeebus.Check(errors.New("boom"))
	return
}

/* TODO: interferes with other tests (because they all run in parallel?)
func TestRun(t *testing.T) {
	go jeebus.Run() // never returns, just makes sure it launches
}
*/

func TestToJson(t *testing.T) {
	expect(t, string(jeebus.ToJson(nil)), `null`)
	expect(t, string(jeebus.ToJson(123)), `123`)
	expect(t, string(jeebus.ToJson("abc")), `"abc"`)
	expect(t, string(jeebus.ToJson([]byte("abc"))), `abc`)
	expect(t, string(jeebus.ToJson(json.RawMessage("abc"))), `abc`)
}

func TestFromToJson(t *testing.T) {
	i := []byte(`{"a":1,"b":[2,"c",3,4.5,{}],"d":null,"e":{"f":"g"}}`)
	o := jeebus.ToJson(jeebus.FromJson(i))
	expect(t, string(i), string(o))
}

func TestDisplayMaxOk(t *testing.T) {
	expect(t, jeebus.DisplayMax("a", 2), "a")
	expect(t, jeebus.DisplayMax("ab", 2), "ab")
	expect(t, jeebus.DisplayMax("abc", 2), "a…")
}