package database

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/jcw/flow"
)

func init() {
	flow.Registry["dump"] = func() flow.Circuitry { return &dumpCmd{} }
	flow.Registry["export"] = func() flow.Circuitry { return &exportCmd{} }
	flow.Registry["import"] = func() flow.Circuitry { return &importCmd{} }
	flow.Registry["get"] = func() flow.Circuitry { return &getCmd{} }
	flow.Registry["put"] = func() flow.Circuitry { return &putCmd{} }
	flow.Registry["keys"] = func() flow.Circuitry { return &keysCmd{} }
}

type dumpCmd struct{ flow.Gadget }

func (g *dumpCmd) Run() {
	odb := openDatabase("")
	defer odb.release()

	odb.iterateOverKeys(flag.Arg(1), flag.Arg(2), func(k string, v []byte) {
		fmt.Printf("%s = %s\n", k, v)
	})
}

type exportCmd struct{ flow.Gadget }

func (g *exportCmd) Run() {
	odb := openDatabase("")
	defer odb.release()

	prefix := flag.Arg(1)
	entries := make(map[string]interface{})

	odb.iterateOverKeys(prefix, "", func(k string, v []byte) {
		var value interface{}
		err := json.Unmarshal(v, &value)
		flow.Check(err)
		key := k[len(prefix):]
		entries[key] = value
	})

	values := make(map[string]map[string]interface{})
	values[prefix] = entries

	s, e := json.MarshalIndent(values, "", "  ")
	flow.Check(e)
	fmt.Println(string(s))
}

type importCmd struct{ flow.Gadget }

func (g *importCmd) Run() {
	data, err := ioutil.ReadFile(flag.Arg(1))
	flow.Check(err)

	var values map[string]map[string]interface{}
	err = json.Unmarshal(data, &values)
	flow.Check(err)

	odb := openDatabase("")
	defer odb.release()

	for prefix, entries := range values {
		var ndel, nadd int

		// get and print all the key/value pairs from the database
		odb.iterateOverKeys(prefix, "", func(k string, v []byte) {
			err = odb.db.Delete([]byte(k), nil)
			flow.Check(err)
			ndel++
		})

		for k, v := range entries {
			val, err := json.Marshal(v)
			flow.Check(err)
			err = odb.db.Put([]byte(prefix+k), val, nil)
			flow.Check(err)
			nadd++
		}

		fmt.Printf("%d deleted, %d added for prefix %q\n", ndel, nadd, prefix)
	}
}

type getCmd struct{ flow.Gadget }

func (g *getCmd) Run() {
	odb := openDatabase("")
	defer odb.release()

	fmt.Println(odb.get(flag.Arg(1)))
}

type putCmd struct{ flow.Gadget }

func (g *putCmd) Run() {
	odb := openDatabase("")
	defer odb.release()

	var value interface{}
	if flag.NArg() > 2 {
		// assume arg is JSON, else pass it in as string
		err := json.Unmarshal([]byte(flag.Arg(2)), &value)
		if err != nil {
			value = flag.Arg(2)
		}
	}
	odb.put(flag.Arg(1), value)
}

type keysCmd struct{ flow.Gadget }

func (g *keysCmd) Run() {
	odb := openDatabase("")
	defer odb.release()

	fmt.Println(strings.Join(odb.keys(flag.Arg(1)), "\n"))
}
