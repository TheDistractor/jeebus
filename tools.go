package jeebus

import (
	"log"
	"time"

	"github.com/codegangsta/cli"
)

func DefineToolCommands() {
	noCustomSubCommands := app == nil || len(app.Commands) == 0

	cmd := AddCommand("run", RunCmd)
	cmd.ShortName = "r"
	cmd.Usage = "launch the web server with messaging and database"

	cmd = AddCommand("subscribe", SubscribeCmd)
	cmd.ShortName = "s"
	cmd.Usage = "subscribe to messages"

	cmd = AddCommand("publish", PublishCmd)
	cmd.ShortName = "p"
	cmd.Usage = "publish a message to a topic"

	cmd = AddCommand("tick", TickCmd)
	cmd.ShortName = "t"
	cmd.Usage = "generate a periodic tick message for testing"

	cmd = AddCommand("dump", DumpCmd)
	cmd.ShortName = "d"
	cmd.Usage = "dump the database contents (offline)"

	cmd = AddCommand("import", ImportCmd)
	cmd.ShortName = "i"
	cmd.Usage = "import a JSON file into the database (offline)"

	cmd = AddCommand("export", ExportCmd)
	cmd.ShortName = "e"
	cmd.Usage = "export the database as JSON (offline)"

	if noCustomSubCommands {
		// also run by default when no other commands have been defined
		app.Action = app.Command("run").Action
	}
}

func RunCmd(c *cli.Context) {
	OpenDatabase()
	StartMessaging()
	RunHttpServer()
}

func SubscribeCmd(c *cli.Context) {
	pattern := c.Args().First()
	if pattern == "" {
		pattern = "#"
	}
	Register(pattern, &SubscribeService{})
	<-StartClient()
}

type SubscribeService struct{}

func (s *SubscribeService) Handle(topic string, payload []byte) {
	log.Println(topic, string(payload))
}

func PublishCmd(c *cli.Context) {
}

func TickCmd(c *cli.Context) {
	topic := c.Args().First()
	if topic == "" {
		topic = "/jb/tick"
	}
	done := StartClient()
	go func() {
		ticker := time.NewTicker(time.Second)
		for tick := range ticker.C {
			Publish(topic, tick.String())
		}
	}()
	<- done
}

func DumpCmd(c *cli.Context) {
}

func ImportCmd(c *cli.Context) {
}

func ExportCmd(c *cli.Context) {
}
