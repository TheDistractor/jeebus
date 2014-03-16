circuits = {}

# simple jeebus setup, with dummy websocket support
circuits.main =
  gadgets: [
    { name: "http", type: "HTTPServer" }
		{ name: "forever", type: "Forever" }
  ]
  feeds: [
    { tag: "/", data: "./app",  to: "http.Handlers" }
    { tag: "/base/", data: "./base",  to: "http.Handlers" }
    { tag: "/common/", data: "./common",  to: "http.Handlers" }
    { tag: "/ws", data: "<websocket>",  to: "http.Handlers" }
    { data: ":3001",  to: "http.Start" }
  ]

# define the websocket handler as just a pipe back to the browser for now
circuits["WebSocket-jeebus"] =
  gadgets: [
    { name: "p", type: "Pipe" }
  ]
  labels: [
    { external: "In", internal: "p.In" }
    { external: "Out", internal: "p.Out" }
  ]

# this app runs a replay simulation with dynamically-loaded decoders
circuits.replay =
  gadgets: [
    { name: "lr", type: "LogReader" }
    { name: "rf", type: "Pipe" } # used to inject an "[RF12demo...]" line
    { name: "w1", type: "LogReplayer" }
    { name: "ts", type: "TimeStamp" }
    { name: "fo", type: "FanOut" }
    { name: "lg", type: "Logger" }
    { name: "st", type: "SketchType" }
    { name: "d1", type: "Dispatcher" }
    { name: "nm", type: "NodeMap" }
    { name: "d2", type: "Dispatcher" }
    { name: "p", type: "Printer" }
  ]
  wires: [
    { from: "lr.Out", to: "w1.In" }
    { from: "rf.Out", to: "ts.In" }
    { from: "w1.Out", to: "ts.In" }
    { from: "ts.Out", to: "fo.In" }
    { from: "fo.Out:lg", to: "lg.In" }
    { from: "fo.Out:st", to: "st.In" }
    { from: "st.Out", to: "d1.In" }
    { from: "d1.Out", to: "nm.In" }
    { from: "nm.Out", to: "d2.In" }
    { from: "d2.Out", to: "p.In" }
  ]
  feeds: [
    { data: "RFg5i2 roomNode boekenkast JC",   to: "nm.Info" }
    { data: "RFg5i3 radioBlip",  to: "nm.Info" }
    { data: "RFg5i4 roomNode washok",   to: "nm.Info" }
    { data: "RFg5i5 roomNode woonkamer",   to: "nm.Info" }
    { data: "RFg5i6 roomNode hal vloer",   to: "nm.Info" }
    { data: "RFg5i9 homePower",  to: "nm.Info" }
    { data: "RFg5i10 roomNode",  to: "nm.Info" }
    { data: "RFg5i11 roomNode logeerkamer",  to: "nm.Info" }
    { data: "RFg5i12 roomNode boekenkast L",  to: "nm.Info" }
    { data: "RFg5i13 roomNode raam halfhoog",  to: "nm.Info" }
    { data: "RFg5i14 otRelay",   to: "nm.Info" }
    { data: "RFg5i15 smaRelay",  to: "nm.Info" }
    { data: "RFg5i18 p1scanner", to: "nm.Info" }
    { data: "RFg5i19 ookRelay",  to: "nm.Info" }
    { data: "RFg5i23 roomNode gang boven",  to: "nm.Info" }
    { data: "RFg5i24 roomNode zolderkamer",  to: "nm.Info" }
    
    { data: "[RF12demo.10] _ i31* g5 @ 868 MHz", to: "rf.In" }
    { data: "./gadgets/rfdata/20121130.txt.gz", to: "lr.Name" }
    { data: "./logger", to: "lg.Dir" }
  ]

# serial port test
circuits.serial =
  gadgets: [
    { name: "sp", type: "SerialPort" }
    { name: "st", type: "SketchType" }
    { name: "d1", type: "Dispatcher" }
    { name: "nm", type: "NodeMap" }
    { name: "d2", type: "Dispatcher" }
  ]
  wires: [
    { from: "sp.From", to: "st.In" }
    { from: "st.Out", to: "d1.In" }
    { from: "d1.Out", to: "nm.In" }
    { from: "nm.Out", to: "d2.In" }
  ]
  feeds: [
    { data: "RFg5i3 radioBlip",  to: "nm.Info" }
    { data: "RFg5i9 homePower",  to: "nm.Info" }
    { data: "RFg5i13 roomNode",  to: "nm.Info" }
    { data: "RFg5i14 otRelay",   to: "nm.Info" }
    { data: "RFg5i15 smaRelay",  to: "nm.Info" }
    { data: "RFg5i18 p1scanner", to: "nm.Info" }
    { data: "RFg5i19 ookRelay",  to: "nm.Info" }
    
    { data: "/dev/tty.usbserial-A901ROSM", to: "sp.Port" }
  ]

# jeeboot server test
circuits.jeeboot =
  gadgets: [
    { name: "sp", type: "SerialPort" }
    { name: "rf", type: "Sketch-RF12demo" }
    { name: "sk", type: "Sink" }
    { name: "jb", type: "JeeBoot" }
  ]
  wires: [
    { from: "sp.From", to: "rf.In" }
    { from: "rf.Out", to: "sk.In" }
    { from: "rf.Rej", to: "sk.In" }
    { from: "rf.Oob", to: "jb.In" }
    { from: "jb.Out", to: "sp.To" }
  ]
  feeds: [
    { data: "/dev/tty.usbserial-A901ROSM", to: "sp.Port" }
  ]

console.log JSON.stringify circuits, null, 4
