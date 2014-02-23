package jeebus

import (
	"log"
	
	"github.com/codegangsta/cli"
)

func DefineToolCommands() {
	noCustomSubCommands := app == nil || len(app.Commands) == 0

	AddCommand(&cli.Command{
		Name:      "run",
		ShortName: "r",
		Usage:     "launch the web server with messaging and database",
		Action: func(c *cli.Context) {
			OpenDatabase()
			StartMessaging()
			RunHttpServer()
		},
	})

	AddCommand(&cli.Command{
		Name:      "subscribe",
		ShortName: "s",
		Usage:     "subscribe to messages",
		Action:    SubscribeCmd,
	})

	AddCommand(&cli.Command{
		Name:      "publish",
		ShortName: "p",
		Usage:     "publish a message to a topic",
		Action:    PublishCmd,
	})

	AddCommand(&cli.Command{
		Name:      "dump",
		ShortName: "d",
		Usage:     "dump the database contents (offline)",
		Action:    DumpCmd,
	})

	AddCommand(&cli.Command{
		Name:      "import",
		ShortName: "i",
		Usage:     "import a JSON file into the database (offline)",
		Action:    ImportCmd,
	})

	AddCommand(&cli.Command{
		Name:      "export",
		ShortName: "e",
		Usage:     "export the database as JSON (offline)",
		Action:    ExportCmd,
	})

	if noCustomSubCommands {
		// also run the default when no other commands have been defined
		app.Action = app.Commands[0].Action
	}
}

func SubscribeCmd(c *cli.Context) {
	pattern := c.Args().First()
	if pattern == "" {
		pattern = "#"
	}
	Register(pattern, &SubscribeService{})
	<- StartClient()
}

type SubscribeService struct{}

func (s *SubscribeService) Handle(topic string, payload []byte) {
	log.Println(topic, string(payload))
}

func PublishCmd(c *cli.Context) {
}

func DumpCmd(c *cli.Context) {
}

func ImportCmd(c *cli.Context) {
}

func ExportCmd(c *cli.Context) {
}
