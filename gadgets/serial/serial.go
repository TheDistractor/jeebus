// +build linux darwin

// Interface to serial port devices (only Linux and Mac OSX).
package serial

import (
	"bufio"
//	"strings"
	"time"

	"github.com/chimera/rs232"
	"github.com/jcw/flow"
)

func init() {
	flow.Registry["SerialPort"] = func() flow.Circuitry { return new(SerialPort) }
	flow.Registry["SketchType"] = func() flow.Circuitry { return new(SketchType) }
}

// Line-oriented serial port, opened once the Port input is set.
type SerialPort struct {
	flow.Gadget
	Port flow.Input
	To   flow.Input
	From flow.Output
}

// Start processing text lines to and from the serial interface.
// Send a bool to adjust RTS or an int to pulse DTR for that many milliseconds.
// Registers as "SerialPort".
func (w *SerialPort) Run() {
	if port, ok := <-w.Port; ok {
		opt := rs232.Options{BitRate: 57600, DataBits: 8, StopBits: 1}
		dev, err := rs232.Open(port.(string), opt)
		flow.Check(err)

		// try to avoid kernel panics due to that wretched buggy FTDI driver!
		// defer func() {
		// 	time.Sleep(time.Second)
		// 	dev.Close()
		// }()
		// time.Sleep(time.Second)

		// separate process to copy data out to the serial port
		go func() {
			for m := range w.To {
				switch v := m.(type) {
				case string:
					dev.Write([]byte(v + "\n"))
				case []byte:
					dev.Write(v)
				case int:
					dev.SetDTR(true) // pulse DTR to reset
					time.Sleep(time.Duration(v) * time.Millisecond)
					dev.SetDTR(false)
				case bool:
					dev.SetRTS(v)
				}
			}
		}()

		scanner := bufio.NewScanner(dev)
		for scanner.Scan() {
			w.From.Send(scanner.Text())
		}
	}
}

