# JeeBus

Experiments with a messaging infrastructure for low-end hardware.

Tested on Mac OSX and BBB so far, but it should also work on other Linux'es.  
Windows will probably require some further adjustment.

## Installation

The Go tools must be installed and GOPATH has to be set, then:

* Lua 5.1 must be installed as library in /usr/lib or /usr/local/lib
* install this project with: `go get -tags luaa github.com/jcw/jeebus/jb`
* go to the source directory using: `cd $GOPATH/github.com/jcw/jeebus`
* launch the server as: `jb run` and leave it running in a console window
* to see messages from MQTT, run `jb see ?topics?` from a separate console
    * `topics` defaults to '#' if not specified, i.e. all topics
* connect a serial port using: `jb serial <dev> <baud> ?tag?`
    * use a `tag` if the serial device does not start by sending a `[...]` line
* to dump all the key/value pairs in the database, run `jb dump`

To make the demo work, the Arduino sketch in `./blinker` has to be uploaded  
to a JeeNode, with a Blink Plug attached to port 1.

Then point the browser to http://localhost:3333/ (or whatever IP the BBB has).
