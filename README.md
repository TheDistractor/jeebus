# JeeBus

Experiments with a messaging infrastructure for low-end hardware.

Tested on Mac OSX and BBB so far, but it should also work on other Linux'es.  
Windows will probably not work as is, due to COM port and DLL differences.

## Installation

The Go tools must be installed and GOPATH has to be set, then:

* Lua 5.1 must be installed as library in /usr/lib or /usr/local/lib
* fetch and install this project with: `go get -tags luaa github.com/jcw/jeebus`
* go to the source directory using: `cd $GOPATH/github.com/jcw/jeebus`
* launch the server as: `jeebus` and leave it running in a console window
* to see all the messages on MQTT, run `jeebus see` from a separate console
* connect a serial port using: `jeebus serial <dev> <baud> ?tag?`
* include a tag if the serial device does not start by sending a `[...]` line

To make the demo work, the Arduino sketch in `./blinker` has to be uploaded  
to a JeeNode, with a Blink Plug attached to port 1.

Then point the browser to http://localhost:3333/ (or whatever IP the BBB has).
