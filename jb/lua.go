package main

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/aarzilli/golua/lua"
	"github.com/jcw/jeebus"
	"github.com/stevedonovan/luar"
)

type LuaDispatchService int

func (s *LuaDispatchService) Handle(m *jeebus.Message) {
	log.Printf("LUA %s %s", m.T, string(m.P))
	split := strings.SplitN(m.T, "/", 2)
	switch split[1] {
	case "register":
		L := newLuaInstance(string(m.P))
		state = L // TODO get rid of this global hack!
		f := luar.NewLuaObjectFromName(L, "service")
		client.Register(string(m.P) + "/#", &LuaRegisteredService{L, f})
	}
}

func newLuaInstance(path string) *lua.State {
	L := lua.NewState()
	L.OpenLibs()
	err := L.DoFile("scripts/" + path + ".lua")
	check(err)
	luar.Register(L, "", luar.Map{
		"publish": client.Publish,
		"dbKeys":  luaDbKeys,
		"dbGet":   luaDbGet,
		"dbSet":   luaDbSet,
	})
	return L
}

type LuaRegisteredService struct {
	L *lua.State
	f *luar.LuaObject
}

var state *lua.State // FIXME hacked, to get it into luaDbGet

func (s *LuaRegisteredService) Handle(m *jeebus.Message) {
	// TODO should auto-reload the Lua script if it has changed on disk
	log.Printf("LUA-SV %s %s", m.T, string(m.P))
	var any interface{}
	err := json.Unmarshal(m.P, &any)
	check(err)
	obj := luar.NewLuaObjectFromValue(s.L, any)
	res, err := s.f.Call(obj)
	check(err)
	log.Printf("RESULT %+v", res)
}

func luaRunWithArgs(args []interface{}) (interface{}, error) {
	if len(args) < 2 {
		log.Fatal("lua needs two or more args:", args)
	}
	path := args[0].(string)
	fname := args[1].(string)
	L := newLuaInstance(path)
	f := luar.NewLuaObjectFromName(L, fname)
	r, e := f.Call(args[2:]...)
	L.Close()
	return r, e
}

func luaDbKeys(prefix string) *luar.LuaObject {
	// TODO look for a way to avoid silly little wrappers like this
	return luar.NewLuaObjectFromValue(state, dbKeys(prefix))
}

func luaDbGet(key string) (obj *luar.LuaObject) {
	v, err := db.Get([]byte(key), nil)
	check(err)
	var any interface{}
	err = json.Unmarshal(v, &any)
	check(err)
	return luar.NewLuaObjectFromValue(state, any)
}

func luaDbSet(key string, value interface{}) {
	if strings.HasPrefix(key, "/") {
		client.Publish(key, value)
		// TODO fall through, i.e. *also* set the value right away
		//  need to inverstigate whether that is a good idea
		//  the alternative causes a slight delay, due to a round trip to MQTT
	}
	if value != nil {
		msg, err := json.Marshal(value)
		check(err)
		db.Put([]byte(key), msg, nil)
	} else {
		db.Delete([]byte(key), nil)
	}
}
