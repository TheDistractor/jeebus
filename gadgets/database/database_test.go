package database

import (
	"os"
	"path"
	"testing"

	"github.com/jcw/flow"
	_ "github.com/jcw/flow/gadgets"
)

var testDb = path.Join(os.TempDir(), "flow-test-db")

func init() {
	println(testDb)
}

func ExampleLevelDB() {
	g := flow.NewCircuit()
	g.Add("db", "LevelDB")
	g.Add("p", "Printer")
	g.Connect("db.Mods", "p.In", 0)
	g.Feed("db.Name", testDb)
	g.Feed("db.In", flow.Tag{"a/b", "123"})
	g.Feed("db.In", flow.Tag{"a/c", "456"})
	g.Feed("db.In", flow.Tag{"<get>", "a/b"})
	g.Feed("db.In", flow.Tag{"<keys>", "a/"})
	g.Feed("db.In", flow.Tag{"a/b", nil})
	g.Feed("db.In", flow.Tag{"<get>", "a/b"})
	g.Feed("db.In", flow.Tag{"<keys>", "a/"})
	g.Feed("db.In", flow.Tag{"a/c", nil})
	g.Run()
	// Output:
	// {Tag:a/b Msg:123}
	// {Tag:a/c Msg:456}
	// Lost flow.Tag: {<get> a/b}
	// Lost string: 123
	// Lost flow.Tag: {<keys> a/}
	// Lost []string: [b c]
	// Lost flow.Tag: {<get> a/b}
	// Lost <nil>: <nil>
	// Lost flow.Tag: {<keys> a/}
	// Lost []string: [c]
	// {Tag:a/b Msg:<nil>}
	// {Tag:a/c Msg:<nil>}
}

func TestDatabase(t *testing.T) {
	g := flow.NewCircuit()
	g.Add("db", "LevelDB")
	g.Feed("db.Name", testDb)
	g.Run()
}
