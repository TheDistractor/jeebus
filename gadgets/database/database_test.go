package database

import (
	"os"
	"path"
	"testing"

	"github.com/jcw/flow"
	_ "github.com/jcw/flow/gadgets"
)

var dbPath = path.Join(os.TempDir(), "flow-test-db")

func init() {
	println(dbPath)
}

func ExampleLevelDB() {
	g := flow.NewCircuit()
	g.Add("db", "LevelDB")
	g.Feed("db.Name", dbPath)
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
	// Lost flow.Tag: {<get> a/b}
	// Lost string: 123
	// Lost flow.Tag: {<keys> a/}
	// Lost []string: [b c]
	// Lost flow.Tag: {<get> a/b}
	// Lost <nil>: <nil>
	// Lost flow.Tag: {<keys> a/}
	// Lost []string: [c]
}

func TestDatabase(t *testing.T) {
	g := flow.NewCircuit()
	g.Add("db", "LevelDB")
	g.Feed("db.Name", dbPath)
	g.Run()
}
