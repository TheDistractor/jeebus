package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aarzilli/golua/lua"
	"github.com/jcw/jeebus"
	"github.com/stevedonovan/luar"
)

func test(L *lua.State) int {
	fmt.Println("hello world! from go!")
	return 0
}

func test2(L *lua.State) int {
	arg := L.CheckInteger(-1)
	argfrombottom := L.CheckInteger(1)
	fmt.Print("test2 arg: ")
	fmt.Println(arg)
	fmt.Print("from bottom: ")
	fmt.Println(argfrombottom)
	return 0
}

func GoFun(args []int) (res map[string]int) {
	res = make(map[string]int)
	for i, val := range args {
		res[strconv.Itoa(i)] = val * val
	}
	return
}

const code = `
	print 'here we go'
	-- Lua tables auto-convert to slices
	local res = GoFun {10,20,30,40}
	-- the result is a map-proxy
	print(111)
	--print(res['1'],res['2'])
	print(222)
	-- which we may explicitly convert to a table
	res = luar.map2table(res)
	for k,v in pairs(res) do
		print(k,v)
	end
`

func setupLua() {
	L := lua.NewState()
	defer L.Close()
	L.OpenLibs()

	// L.Register("test2", test2)

	L.GetField(lua.LUA_GLOBALSINDEX, "print")
	L.PushString("Hello World!")
	L.Call(1, 0)

	L.PushGoFunction(test)

	L.PushGoFunction(test2)
	L.PushInteger(42)
	L.Call(1, 0)

	L.Call(0, 0) // hello world! from go!

	luar.Register(L, "", luar.Map{
		"Print": fmt.Println,
		"MSG":   "hello", // can also register constants
		"GoFun": GoFun,
	})

	err := L.DoString(code)
	fmt.Printf("error %v\n", err)
}

type LuaDispatchService int

func (s *LuaDispatchService) Handle(m *jeebus.Message) {
	log.Printf("LUA %s %s", m.T, string(m.P))
	split := strings.SplitN(m.T, "/", 2)
	switch split[1] {
	case "register":
		L := lua.NewState()
		L.OpenLibs()
		err := L.DoFile("scripts/" + string(m.P) + ".lua")
		check(err)
		f := luar.NewLuaObjectFromName(L, "service")
		luar.Register(L, "", luar.Map{
			"publish": jeebus.Publish,
			"dbKeys":  luaDbKeys,
			"dbGet":   luaDbGet,
			"dbSet":   luaDbSet,
		})
		// FIXME assumes path is "sv/..."
		state = L
		svClient.Register(string(m.P)[3:], &LuaRegisteredService{L, f})
	}
}

type LuaRegisteredService struct {
	L *lua.State
	f *luar.LuaObject
}

var state *lua.State // FIXME hacked, to get it into luaDbGet

func (s *LuaRegisteredService) Handle(m *jeebus.Message) {
	// TODO should auto-reload the Lua script if it has changed on disk
	log.Printf("LUA-RS %s %s", m.T, string(m.P))
	var any interface{}
	err := json.Unmarshal(m.P, &any)
	check(err)
	obj := luar.NewLuaObjectFromValue(s.L, any)
	res, err := s.f.Call(obj)
	check(err)
	log.Printf("RESULT %+v", res)
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
		log.Fatal("cannot use dbSet, must publish: ", key)
	}
	if value != nil {
		msg, err := json.Marshal(value)
		check(err)
		db.Put([]byte(key), msg, nil)
	} else {
		db.Delete([]byte(key), nil)
	}
}
