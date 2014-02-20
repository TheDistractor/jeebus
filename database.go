package jeebus

import (
	"bytes"

	"github.com/syndtr/goleveldb/leveldb"
	// "github.com/syndtr/goleveldb/leveldb/opt"
)

var (
	db *leveldb.DB
)

type DatabaseService struct{}

func (s *DatabaseService) Handle(topic string, payload interface{}) {
	Put(topic, payload)
}

func OpenDatabase() error {
	// o := &opt.Options{ ErrorIfMissing: true }
	var err error
	db, err = leveldb.OpenFile(Settings.DbDir, nil)
	if err == nil {
		Register("/#", &DatabaseService{})
	}
	return err
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
	from, to, skip := []byte(prefix), []byte(prefix+"~"), len(prefix)
	// from, to, skip := []byte(prefix+"/"), []byte(prefix+"/~"), len(prefix)+1
	prev := []byte("/") // impossible value, this never matches actual results

	iter := db.NewIterator(nil)
	defer iter.Release()

	iter.Seek(from)
	for iter.Valid() {
		k := iter.Key()
		// fmt.Printf(" -> %s = %s\n", k, iter.Value())
		if bytes.Compare(k, to) > 0 {
			break
		}
		i := bytes.IndexRune(k[skip:], '/') + skip
		if i < skip {
			i = len(k)
		}
		// fmt.Printf(" DK %d %d %d %s %s\n", skip, len(prev), i, prev, k)
		if !bytes.Equal(prev, k[skip:i]) {
			// need to make a copy of the key, since it's owned by iter
			prev = make([]byte, i-skip)
			copy(prev, k[skip:i])
			// fmt.Printf("ADD %s\n", prev)
			results = append(results, string(prev))
		}
		if !iter.Next() {
			break
		}
	}
	return
}
