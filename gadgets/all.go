// Convenience package to wrap all the gadgets available in JeeBus.
package jeebus

import (
	_ "github.com/jcw/jeebus/gadgets/database"
	_ "github.com/jcw/jeebus/gadgets/decoders"
	_ "github.com/jcw/jeebus/gadgets/javascript"
	_ "github.com/jcw/jeebus/gadgets/network"
	_ "github.com/jcw/jeebus/gadgets/rfdata"
	_ "github.com/jcw/jeebus/gadgets/serial"
)

var Version = "0.9.0"
