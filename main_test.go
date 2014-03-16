package main

import "github.com/jcw/flow"

func Example() {
	c := flow.NewCircuit()
	c.Add("clock", "Clock")
	c.Add("counter", "Counter") // will return 0 when not hooked up
	c.Add("pipe", "Pipe")
	c.Add("printer", "Printer")
	c.Add("repeater", "Repeater")
	c.Add("serial", "SerialPort")
	c.Add("sink", "Sink")
	c.Add("timer", "Timer")
	c.Add("timestamp", "TimeStamp")
	c.Run()
	// Output:
	// Lost int: 0
}
