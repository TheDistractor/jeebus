package rfdata

import (
	"github.com/jcw/flow"
	_ "github.com/jcw/flow/gadgets"
)

func ExampleCalcCrc16() {
	g := flow.NewCircuit()
	g.Add("c", "CalcCrc16")
	g.Feed("c.In", []byte("abc"))
	g.Run()
	// Output:
	// Lost []uint8: [97 98 99]
	// Lost flow.Tag: {<crc16> 22345}
}

func ExampleReadTextFile() {
	g := flow.NewCircuit()
	g.Add("r", "ReadTextFile")
	g.Add("c", "Counter")
	g.Connect("r.Out", "c.In", 0)
	g.Feed("r.In", "./blinkAvr1.hex")
	g.Run()
	// Output:
	// Lost flow.Tag: {<open> ./blinkAvr1.hex}
	// Lost flow.Tag: {<close> ./blinkAvr1.hex}
	// Lost int: 47
}

func ExampleIntelHexToBin() {
	g := flow.NewCircuit()
	g.Add("r", "ReadTextFile")
	g.Add("b", "IntelHexToBin")
	g.AddCircuitry("n", flow.Transformer(func(m flow.Message) flow.Message {
		if v, ok := m.([]byte); ok {
			m = len(v)
		}
		return m
	}))
	g.Connect("r.Out", "b.In", 0)
	g.Connect("b.Out", "n.In", 0)
	g.Feed("r.In", "./blinkAvr1.hex")
	g.Run()
	// Output:
	// Lost flow.Tag: {<open> ./blinkAvr1.hex}
	// Lost flow.Tag: {<addr> 0}
	// Lost int: 726
	// Lost flow.Tag: {<close> ./blinkAvr1.hex}
}

func ExampleBinaryFill() {
	g := flow.NewCircuit()
	g.Add("f", "BinaryFill")
	g.Feed("f.In", []byte("abcdef"))
	g.Feed("f.Len", 5)
	g.Run()
	// Output:
	// Lost []uint8: [97 98 99 100 101 102 255 255 255 255]
}

func ExampleDelay() {
	g := flow.NewCircuit()
	g.Add("d", "Delay")
	g.Add("p", "Printer")
	g.Feed("d.Delay", "10")
	g.Feed("d.In", "abc")
	g.Feed("p.In", "def")
	g.Run()
	// Output:
	// string: def
	// Lost string: abc
}
