package jeebus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/codegangsta/cli"
)

func DefineToolCommands() {
	cmd := AddCommand("run", app.Action) // alias for default case
	cmd.ShortName = "r"
	cmd.Usage = "launch the web server with messaging and database (default)"

	cmd = AddCommand("subscribe", SubscribeCmd)
	cmd.ShortName = "s"
	cmd.Usage = "subscribe to messages"

	cmd = AddCommand("publish", PublishCmd)
	cmd.ShortName = "p"
	cmd.Usage = "publish a message to a topic"

	cmd = AddCommand("tick", TickCmd)
	cmd.ShortName = "t"
	cmd.Usage = "generate a periodic tick message for testing"

	cmd = AddCommand("dump", DumpCmd)
	cmd.ShortName = "d"
	cmd.Usage = "dump the database contents (offline)"

	cmd = AddCommand("import", func(c *cli.Context) {
		if len(c.Args()) < 1 {
			log.Fatalf("usage: jb import <jsonfile>")
		}
		OpenDatabase()
		ImportJsonData(c.Args().First())
	})
	cmd.ShortName = "i"
	cmd.Usage = "import a JSON file into the database (offline)"

	cmd = AddCommand("export", func(c *cli.Context) {
		if len(c.Args()) < 1 {
			log.Fatalf("usage: jb export <prefix>")
		}
		OpenDatabase()
		ExportJsonData(c.Args().First())
	})
	cmd.ShortName = "e"
	cmd.Usage = "export the database as JSON (offline)"
}

func SubscribeCmd(c *cli.Context) {
	pattern := c.Args().First()
	if pattern == "" {
		pattern = "#"
	}
	Register(pattern, &SubscribeService{})
	<-StartClient()
}

type SubscribeService struct{}

func (s *SubscribeService) Handle(topic string, payload []byte) {
	log.Println(topic, string(payload))
}

func PublishCmd(c *cli.Context) {
	if len(c.Args()) < 2 {
		log.Fatalf("usage: jb pub <topic> ?<jsonval>?")
	}
	topic, value := c.Args().Get(0), c.Args().Get(1)
	StartClient()
	Publish(topic, []byte(value))
	// TODO: need to close gracefully, and not too soon!
	time.Sleep(10 * time.Millisecond)
}

func TickCmd(c *cli.Context) {
	topic := c.Args().First()
	if topic == "" {
		topic = "/jb/tick"
	}
	done := StartClient()
	go func() {
		ticker := time.NewTicker(time.Second)
		for tick := range ticker.C {
			Publish(topic, tick.String())
		}
	}()
	<-done
}

func DumpCmd(c *cli.Context) {
	from, to := c.Args().Get(0), c.Args().Get(1)
	if to == "" {
		to = from + "~" // FIXME this assumes all key chars are less than "~"
	}

	OpenDatabase()
	// get and print all the key/value pairs from the database
	iter := db.NewIterator(nil)
	iter.Seek([]byte(from))
	for iter.Valid() {
		if string(iter.Key()) > to {
			break
		}
		fmt.Printf("%s = %s\n", iter.Key(), iter.Value())
		if !iter.Next() {
			break
		}
	}
	iter.Release()
}

func ImportJsonData(filename string) {
	data, err := ioutil.ReadFile(filename)
	Check(err)

	var values map[string]map[string]*json.RawMessage
	err = json.Unmarshal(data, &values)
	Check(err)

	for prefix, entries := range values {
		limit := prefix + "~" // FIXME see below, same as for dumpDatabase()
		var ndel, nadd int

		// get and print all the key/value pairs from the database
		iter := db.NewIterator(nil)
		iter.Seek([]byte(prefix))
		for iter.Valid() {
			key := string(iter.Key())
			if key > limit {
				break
			}
			err = db.Delete([]byte(key), nil)
			Check(err)
			ndel++
			if !iter.Next() {
				break
			}
		}
		iter.Release()

		for k, v := range entries {
			err = db.Put([]byte(prefix+k), *v, nil)
			Check(err)
			nadd++
		}

		fmt.Printf("%d deleted, %d added for prefix %q\n", ndel, nadd, prefix)
	}
}

func ExportJsonData(prefix string) {
	limit := prefix + "~" // FIXME see below, same as for dumpDatabase()
	entries := make(map[string]interface{})

	// get and print all the key/value pairs from the database
	iter := db.NewIterator(nil)
	iter.Seek([]byte(prefix))
	for iter.Valid() {
		key := iter.Key()[len(prefix):]
		if string(iter.Key()) > limit {
			break
		}
		var value interface{}
		err := json.Unmarshal(iter.Value(), &value)
		Check(err)
		entries[string(key)] = value
		if !iter.Next() {
			break
		}
	}
	iter.Release()

	values := make(map[string]map[string]interface{})
	values[prefix] = entries

	s, e := json.MarshalIndent(values, "", "  ")
	Check(e)
	fmt.Println(string(s))
}
