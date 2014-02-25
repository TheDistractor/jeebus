package jeebus

import (
	"log"
	"strings"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	// dbopt "github.com/syndtr/goleveldb/leveldb/opt"
	dbutil "github.com/syndtr/goleveldb/leveldb/util"
)

var (
	db        *leveldb.DB
	dbStarted bool
	attached  = make(map[string]map[string]int) // prefix -> tag -> refcount
)

type DatabaseService struct{}

func (s *DatabaseService) Handle(topic string, payload []byte) {
	Put(topic, payload)

	// send out websocket messages for all matching attached topics
	msg := make(map[string]interface{})
	msg[topic] = FromJson(payload) // TODO: unfortunate extra decode/encode
	json := ToJson(msg)

	for k, v := range attached {
		if strings.HasPrefix(topic, k) {
			for dest, _ := range v {
				log.Println("dispatch", dest, string(payload))
				Dispatch("ws/"+dest, json) // direct dispatch, no MQTT
			}
		}
	}
}

func OpenDatabase() {
	if dbStarted {
		return
	}
	dbStarted = true

	// o := &opt.Options{ ErrorIfMissing: true }
	var err error
	db, err = leveldb.OpenFile(Settings.DbDir, nil)
	Check(err)

	// no need to publish these, since messaging hasn't been started up yet
	Put("/jb/info", map[string]interface{}{
		"started": time.Now().Format(time.RFC822Z),
		"version": Version,
	})
	Put("/jb/settings", Settings)

	Register("/#", &DatabaseService{})

	Define("db-get", func(orig string, args []interface{}) interface{} {
		return Get(args[0].(string))
	})
	Define("db-keys", func(orig string, args []interface{}) interface{} {
		return Keys(args[0].(string))
	})
	Define("attach", attachRpc)
	Define("detach", detachRpc)
}

func attachRpc(orig string, args []interface{}) interface{} {
	prefix := args[0].(string)
	if _, ok := attached[prefix]; !ok {
		attached[prefix] = make(map[string]int)
	}
	if _, ok := attached[prefix][orig]; !ok {
		attached[prefix][orig] = 0
	}
	attached[prefix][orig]++

	if Settings.VerboseRpc {
		log.Println("attached", prefix, orig)
	}

	result := make(map[string]interface{})
	IterateOverKeys(prefix, "", func(k string, v []byte) {
		result[k] = FromJson(v)
	})
	return result
}

func detachRpc(orig string, args []interface{}) interface{} {
	prefix := args[0].(string)
	if v, ok := attached[prefix]; ok {
		if _, ok := v[orig]; ok {
			attached[prefix][orig]--
			if attached[prefix][orig] <= 0 {
				delete(attached[prefix], orig)
				if len(attached[prefix]) == 0 {
					delete(attached, prefix)
				}
			}
		}
	}

	if Settings.VerboseRpc {
		log.Println("detached", prefix, orig)
	}

	return nil
}

func Get(key string) interface{} {
	value, err := db.Get([]byte(key), nil)
	if err == leveldb.ErrNotFound {
		return nil
	}
	Check(err)
	return FromJson(value)
}

func Put(key string, value interface{}) {
	if value != nil {
		db.Put([]byte(key), ToJson(value), nil)
	} else {
		db.Delete([]byte(key), nil)
	}
}

func Keys(prefix string) (results []string) {
	// TODO: decide whether this key logic is the most useful & least confusing
	// TODO: should use skips and reverse iterators once the db gets larger!
	skip := len(prefix)
	prev := "/" // impossible value, this never matches actual results

	IterateOverKeys(prefix, "", func(k string, v []byte) {
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

func IterateOverKeys(from, to string, iterFun func(string, []byte)) {
	slice := &dbutil.Range{[]byte(from), []byte(to)}
	if len(to) == 0 {
		slice.Limit = append(slice.Start, 0xFF)
	}

	iter := db.NewIterator(slice, nil)
	defer iter.Release()

	for iter.Next() {
		iterFun(string(iter.Key()), iter.Value())
	}
}
