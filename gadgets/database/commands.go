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
	flow.Registry["dbdump"] = func() flow.Circuitry { return new(dumpCmd) }
	flow.Registry["dbexport"] = func() flow.Circuitry { return new(exportCmd) }
	flow.Registry["dbimport"] = func() flow.Circuitry { return new(importCmd) }
	flow.Registry["dbget"] = func() flow.Circuitry { return new(getCmd) }
	flow.Registry["dbput"] = func() flow.Circuitry { return new(putCmd) }
	flow.Registry["dbkeys"] = func() flow.Circuitry { return new(keysCmd) }
}

type dumpCmd struct{ flow.Gadget }

func (g *dumpCmd) Run() {
	openDatabase()

	dbIterateOverKeys(flag.Arg(1), flag.Arg(2), func(k string, v []byte) {
		fmt.Printf("%s = %s\n", k, v)
	})
}

type exportCmd struct{ flow.Gadget }

func (g *exportCmd) Run() {
	openDatabase()

	prefix := flag.Arg(1)
	entries := make(map[string]interface{})

	dbIterateOverKeys(prefix, "", func(k string, v []byte) {
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

	openDatabase()

	for prefix, entries := range values {
		var ndel, nadd int

		// get and print all the key/value pairs from the database
		dbIterateOverKeys(prefix, "", func(k string, v []byte) {
			err = db.Delete([]byte(k), nil)
			flow.Check(err)
			ndel++
		})

		for k, v := range entries {
			val, err := json.Marshal(v)
			flow.Check(err)
			err = db.Put([]byte(prefix+k), val, nil)
			flow.Check(err)
			nadd++
		}

		fmt.Printf("%d deleted, %d added for prefix %q\n", ndel, nadd, prefix)
	}
}

type getCmd struct{ flow.Gadget }

func (g *getCmd) Run() {
	openDatabase()

	fmt.Println(dbGet(flag.Arg(1)))
}

type putCmd struct{ flow.Gadget }

func (g *putCmd) Run() {
	openDatabase()

	var value interface{}
	if flag.NArg() > 2 {
		// assume arg is JSON, else pass it in as string
		err := json.Unmarshal([]byte(flag.Arg(2)), &value)
		if err != nil {
			value = flag.Arg(2)
		}
	}
	dbPut(flag.Arg(1), value)
}

type keysCmd struct{ flow.Gadget }

func (g *keysCmd) Run() {
	openDatabase()

	fmt.Println(strings.Join(dbKeys(flag.Arg(1)), "\n"))
}
