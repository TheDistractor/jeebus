package database

import (
	"os"
	"path"
	"testing"

	"github.com/jcw/flow"
	_ "github.com/jcw/flow/gadgets"
)

func init() {
	testDb := path.Join(os.TempDir(), "flow-test-db")
	println(testDb)
	flow.Config["DATA_DIR"] = testDb
}

func ExampleLevelDB() {
	g := flow.NewCircuit()
	g.Add("db", "LevelDB")
	g.Feed("db.In", flow.Tag{"a/b", "123"})
	g.Feed("db.In", flow.Tag{"a/c", "456"})
	g.Feed("db.In", flow.Tag{"<get>", "a/b"})
	g.Feed("db.In", flow.Tag{"<range>", "a/"})
	g.Feed("db.In", flow.Tag{"<keys>", "a/"})
	g.Feed("db.In", flow.Tag{"a/b", nil})
	g.Feed("db.In", flow.Tag{"<get>", "a/b"})
	g.Feed("db.In", flow.Tag{"<keys>", "a/"})
	g.Feed("db.In", flow.Tag{"a/c", nil})
	g.Run()
	// Output:
	// Lost flow.Tag: {a/b 123}
	// Lost flow.Tag: {a/c 456}
	// Lost flow.Tag: {<get> a/b}
	// Lost string: 123
	// Lost flow.Tag: {<range> a/}
	// Lost flow.Tag: {a/b 123}
	// Lost flow.Tag: {a/c 456}
	// Lost flow.Tag: {<keys> a/}
	// Lost string: b
	// Lost string: c
	// Lost flow.Tag: {a/b <nil>}
	// Lost flow.Tag: {<get> a/b}
	// Lost <nil>: <nil>
	// Lost flow.Tag: {<keys> a/}
	// Lost string: c
	// Lost flow.Tag: {a/c <nil>}
}

func TestDatabase(t *testing.T) {
	g := flow.NewCircuit()
	g.Add("db", "LevelDB")
	g.Run()
}
