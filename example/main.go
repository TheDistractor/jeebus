package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/jcw/jeebus"
)

func main() {
	// TODO: this message is needed for testing, which picks up the pid from it
	println("JeeBus example", jeebus.Version, "pid", os.Getpid())

	app := jeebus.NewApp("example", jeebus.Version)
	app.Usage = "a minimal application based on JeeBus"

	cmd := jeebus.AddCommand("foo", func(c *cli.Context) {
		println("bar")
	})
	cmd.Usage = "dummy command"

	jeebus.Run()
}
