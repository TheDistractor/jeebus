package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/jcw/jeebus"
)

func init() {
	log.SetFlags(log.Ltime)
}

func main() {
	// TODO: this message is needed for testing, which picks up the pid from it
	println("JeeBus example", jeebus.Version, "pid", os.Getpid())

	app := jeebus.NewApp("example", jeebus.Version)
	app.Usage = "a minimal application based on JeeBus"

	jeebus.NewCommand(&cli.Command{
		Name:  "foo",
		Usage: "dummy command",
		Action: func(c *cli.Context) {
			println("bar")
		},
	})

	jeebus.Run()
}
