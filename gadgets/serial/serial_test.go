package serial

import (
	"testing"

	"github.com/jcw/flow"
	_ "github.com/jcw/flow/gadgets"
)

func TestSerial(t *testing.T) {
	t.Skip("skipping serial test, never ends and needs hardware.")
	// The following example never ends, comment out the above to try it out
	g := flow.NewCircuit()
	g.Add("s", "SerialPort")
	g.Add("t", "SketchType")
	g.Add("p", "Printer")
	g.Connect("s.From", "t.In", 0)
	g.Connect("t.Out", "p.In", 0)
	g.Feed("s.Port", "/dev/tty.usbserial-A900ad5m")
	g.Run()
}
