# JeeBus

Experiments with a messaging infrastructure for low-end hardware.

Tested on Mac OSX and BBB so far, but it should also work on other Linux'es.  
Windows will probably not work as is, due to COM port and DLL differences.

## Installation

The Go tools must be installed and GOPATH has to be set, then:

* A recent LevelDB must be installed, i.e. header files and the shared library.
* Lua 5.1 must be installed as shared library, a static build is not enough.
* Fetch and install this project with: `go get github.com/jcw/jeebus`
* Launch as: `jeebus ?serial-device?`
* On BBB: `LD_LIBRARY_PATH=/usr/local/lib jeebus ?serial-device?`

To make the demo work, the Arduino sketch in `./blinker` has to be uploaded  
to a JeeNode, with a Blink Plug attached to port 1.

Then point the browser to http://localhost:3333/ (or whatever IP the BBB has).
