// Interface to the LevelDB database.
package database

import (
	"encoding/json"
	"flag"
	"strings"
	"sync"

	"github.com/golang/glog"
	"github.com/jcw/flow"
	"github.com/syndtr/goleveldb/leveldb"
	dbutil "github.com/syndtr/goleveldb/leveldb/util"
)

var dbPath = flag.String("db", "./data", "location of the LevelDB database")

func init() {
	flow.Registry["LevelDB"] = func() flow.Circuitry { return &LevelDB{} }
}

var (
	dbMutex sync.Mutex
	dbMap   = map[string]*openDb{}
)

type openDb struct {
	name string
	db   *leveldb.DB
	refs int
}

func (odb *openDb) release() {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	odb.refs--
	if odb.refs <= 0 {
		odb.db.Close()
		delete(dbMap, odb.name)
	}
}

func (w *openDb) iterateOverKeys(from, to string, fun func(string, []byte)) {
	slice := &dbutil.Range{[]byte(from), []byte(to)}
	if len(to) == 0 {
		slice.Limit = append(slice.Start, 0xFF)
	}

	iter := w.db.NewIterator(slice, nil)
	defer iter.Release()

	for iter.Next() {
		fun(string(iter.Key()), iter.Value())
	}
}

func (w *openDb) get(key string) (any interface{}) {
	glog.Infoln("get", key)
	data, err := w.db.Get([]byte(key), nil)
	if err == leveldb.ErrNotFound {
		return nil
	}
	flow.Check(err)
	err = json.Unmarshal(data, &any)
	flow.Check(err)
	return
}

func (w *openDb) put(key string, value interface{}) {
	glog.Infoln("put", key, value)
	if value != nil {
		data, err := json.Marshal(value)
		flow.Check(err)
		w.db.Put([]byte(key), data, nil)
	} else {
		w.db.Delete([]byte(key), nil)
	}
}

func (w *openDb) keys(prefix string) (results []string) {
	glog.Infoln("keys", prefix)
	// TODO: decide whether this key logic is the most useful & least confusing
	// TODO: should use skips and reverse iterators once the db gets larger!
	skip := len(prefix)
	prev := "/" // impossible value, this never matches actual results

	w.iterateOverKeys(prefix, "", func(k string, v []byte) {
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

func openDatabase(name string) *openDb {
	if name == "" {
		name = *dbPath
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()

	odb, ok := dbMap[name]
	if !ok {
		db, err := leveldb.OpenFile(name, nil)
		flow.Check(err)
		odb = &openDb{name, db, 0}
		dbMap[name] = odb
	}
	odb.refs++
	return odb
}

// LevelDB is a multi-purpose .Feed( to get, put, and scan keys in a database.
// Acts on tags received on the input port. Registers itself as "LevelDB".
type LevelDB struct {
	flow.Gadget
	Name flow.Input
	In   flow.Input
	Out  flow.Output
	Mods flow.Output

	odb *openDb
}

// Open the database and start listening to incoming get/put/keys requests.
func (w *LevelDB) Run() {
	// if a name is given, use it, else use the default from the command line
	name := ""
	if m, ok := <-w.Name; ok {
		name = m.(string)
	}
	w.odb = openDatabase(name)
	defer w.odb.release()
	for m := range w.In {
		if tag, ok := m.(flow.Tag); ok {
			switch tag.Tag {
			case "<get>":
				key := tag.Msg.(string)
				w.Out.Send(m)
				w.Out.Send(w.odb.get(key))
			case "<keys>":
				prefix := tag.Msg.(string)
				w.Out.Send(m)
				w.Out.Send(w.odb.keys(prefix))
			default:
				w.odb.put(tag.Tag, tag.Msg)
				w.Mods.Send(tag)
			}
		} else {
			w.Out.Send(m)
		}
	}
}
