// Interface to the LevelDB database.
package database

// glog levels:
//	1 = changes to registry
//  2 = changes to database
//  3 = database access

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/golang/glog"
	"github.com/jcw/flow"
	"github.com/syndtr/goleveldb/leveldb"
	dbutil "github.com/syndtr/goleveldb/leveldb/util"
)

var (
	once sync.Once
	db   openDb
)

type openDb struct{ *leveldb.DB }

func init() {
	flow.Registry["LevelDB"] = func() flow.Circuitry { return new(LevelDB) }
}

func dbIterateOverKeys(from, to string, fun func(string, []byte)) {
	slice := &dbutil.Range{[]byte(from), []byte(to)}
	if len(to) == 0 {
		slice.Limit = append(slice.Start, 0xFF)
	}

	iter := db.NewIterator(slice, nil)
	defer iter.Release()

	for iter.Next() {
		fun(string(iter.Key()), iter.Value())
	}
}

func dbGet(key string) (any interface{}) {
	glog.V(3).Infoln("get", key)
	data, err := db.Get([]byte(key), nil)
	if err == leveldb.ErrNotFound {
		return nil
	}
	flow.Check(err)
	err = json.Unmarshal(data, &any)
	flow.Check(err)
	return
}

func dbPut(key string, value interface{}) {
	glog.V(2).Infoln("put", key, value)
	if value != nil {
		data, err := json.Marshal(value)
		flow.Check(err)
		db.Put([]byte(key), data, nil)
	} else {
		db.Delete([]byte(key), nil)
	}
}

func dbKeys(prefix string) (results []string) {
	glog.V(3).Infoln("keys", prefix)
	// TODO: decide whether this key logic is the most useful & least confusing
	// TODO: should use skips and reverse iterators once the db gets larger!
	skip := len(prefix)
	prev := "/" // impossible value, this never matches actual results

	dbIterateOverKeys(prefix, "", func(k string, v []byte) {
		i := strings.IndexRune(k[skip:], '/') + skip
		if i < skip {
			i = len(k)
		}
		if prev != k[skip:i] {
			// need to make a copy of the key, since it's owned by iter
			prev = k[skip:i]
			results = append(results, string(prev))
		}
	})
	return
}

func dbRegister(key string) {
	data, err := db.Get([]byte(key), nil)
	if err == leveldb.ErrNotFound {
		glog.Warningln("cannot register:", key)
		return
	}
	name := key[strings.LastIndex(key, "/")+1:]
	glog.V(1).Infof("register %s: %d bytes (%s)", name, len(data), key)
	flow.Registry[key] = func() flow.Circuitry {
		c := flow.NewCircuit()
		c.LoadJSON(data)
		return c
	}
}

func openDatabase() {
	// opening the database takes time, make sure we don't re-enter this code
	once.Do(func() {
		dbPath := flow.Config["DATA_DIR"]
		if dbPath == "" {
			glog.Fatalln("cannot open database, DATA_DIR not set")
		}
		d, err := leveldb.OpenFile(dbPath, nil)
		flow.Check(err)
		db = openDb{d}
	})
}

// Get an entry from the database, returns nil if not found.
func Get(key string) interface{} {
	openDatabase()
	return dbGet(key)
}

// Store or delete an entry in the database.
func Put(key string, value interface{}) {
	openDatabase()
	dbPut(key, value)
}

// Get a list of keys from the database, given a prefix.
func Keys(prefix string) []string {
	openDatabase()
	return dbKeys(prefix)
}

// LevelDB is a multi-purpose gadget to get, put, and scan keys in a database.
// Acts on tags received on the input port. Registers itself as "LevelDB".
type LevelDB struct {
	flow.Gadget
	In   flow.Input
	Out  flow.Output
	Mods flow.Output
}

// Open the database and start listening to incoming get/put/keys requests.
func (w *LevelDB) Run() {
	openDatabase()
	for m := range w.In {
		if tag, ok := m.(flow.Tag); ok {
			switch tag.Tag {
			case "<get>":
				w.Out.Send(m)
				w.Out.Send(dbGet(tag.Msg.(string)))
			case "<keys>":
				w.Out.Send(m)
				for _, s := range dbKeys(tag.Msg.(string)) {
					w.Out.Send(s)
				}
			case "<clear>":
				prefix := tag.Msg.(string)
				glog.V(2).Infoln("clear", prefix)
				dbIterateOverKeys(prefix, "", func(k string, v []byte) {
					db.Delete([]byte(k), nil)
				})
				w.Mods.Send(m)
			case "<range>":
				prefix := tag.Msg.(string)
				glog.V(3).Infoln("range", prefix)
				w.Out.Send(m)
				dbIterateOverKeys(prefix, "", func(k string, v []byte) {
					var any interface{}
					err := json.Unmarshal(v, &any)
					flow.Check(err)
					w.Out.Send(flow.Tag{k, any})
				})
			case "<register>":
				dbRegister(tag.Msg.(string))
				w.Mods.Send(m)
			default:
				if strings.HasPrefix(tag.Tag, "<") {
					w.Out.Send(m) // pass on other tags without processing
				} else {
					dbPut(tag.Tag, tag.Msg)
					w.Mods.Send(m)
				}
			}
		} else {
			w.Out.Send(m)
		}
	}
}
