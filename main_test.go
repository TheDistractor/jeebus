package main

import "github.com/jcw/flow"

func Example() {
	g := flow.NewCircuit()
	g.Add("clock", "Clock")
	g.Add("counter", "Counter") // will return 0 when not hooked up
	g.Add("pipe", "Pipe")
	g.Add("printer", "Printer")
	g.Add("repeater", "Repeater")
	g.Add("serial", "SerialPort")
	g.Add("sink", "Sink")
	g.Add("timer", "Timer")
	g.Add("timestamp", "TimeStamp")
	g.Run()
	// Output:
	// Lost int: 0
}
