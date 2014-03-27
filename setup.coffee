#!/usr/bin/env coffee

circuits = {}

# jeebus setup, stores in db and publishes on mqtt
circuits.main =
  gadgets: [
    { name: "http", type: "HTTPServer" }
    { name: "init", type: "init" }
  ]
  feeds: [
    { tag: "/", data: "./app",  to: "http.Handlers" }
    { tag: "/base/", data: "./base",  to: "http.Handlers" }
    { tag: "/ws", data: "<websocket>",  to: "http.Handlers" }
    { data: ":3000",  to: "http.Port" }
  ]

# init circuit for HouseMon, which starts its own http server.
circuits.init =
  gadgets: [
    { name: "mqtt", type: "MQTTServer" }
		{ name: "replay", type: "replay" }
    { name: "pub", type: "MQTTPub" }
    { name: "dummy", type: "Pipe" } # needed for dispatcher in HouseMon
    { name: "driverFill", type: "driverFill" } # pre-load the database
    { name: "tableFill", type: "tableFill" }   # pre-load the database
  ]
  wires: [
    { from: "replay.Out", to: "pub.In" }
    { from: "mqtt.PortOut", to: "pub.Port" }
  ]
  feeds: [
    { data: ":1883",  to: "mqtt.Port" }
  ]
  labels: [
    { external: "In", internal: "dummy.In" }
    { external: "Out", internal: "dummy.Out" }
  ]

# define the websocket handler using a loop in and out of RpcHandler
circuits["WebSocket-jeebus"] =
  gadgets: [
    { name: "rpc", type: "RpcHandler" }
  ]
  labels: [
    { external: "In", internal: "rpc.In" }
    { external: "Out", internal: "rpc.Out" }
  ]

# this app runs a replay simulation with dynamically-loaded decoders
circuits.replay =
  gadgets: [
    { name: "lr", type: "LogReader" }
    { name: "rf", type: "Pipe" } # used to inject an "[RF12demo...]" line
    { name: "w1", type: "LogReplayer" }
    { name: "ts", type: "TimeStamp" }
    { name: "f1", type: "FanOut" }
    { name: "lg", type: "Logger" }
    { name: "db", type: "rf12toDatabase" }
  ]
  wires: [
    { from: "lr.Out", to: "w1.In" }
    { from: "rf.Out", to: "ts.In" }
    { from: "w1.Out", to: "ts.In" }
    { from: "ts.Out", to: "f1.In" }
    { from: "f1.Out:lg", to: "lg.In" }
    { from: "f1.Out:db", to: "db.In" }
  ]
  feeds: [
    { data: "[RF12demo.10] _ i31* g5 @ 868 MHz", to: "rf.In" }
    { data: "./gadgets/rfdata/20121130.txt.gz", to: "lr.Name" }
    { data: "./logger", to: "lg.Dir" }
  ]
  labels: [
    { external: "Out", internal: "db.Out" }
  ]
  
# the node mapping for nodes at JeeLabs, as pre-configured circuit
circuits.nodesJeeLabs =
  gadgets: [
    { name: "nm", type: "NodeMap" }
  ]
  feeds: [
    { data: "RFg5i2,roomNode,boekenkast JC",  to: "nm.Info" }
    { data: "RFg5i3,radioBlip,werkkamer",     to: "nm.Info" }
    { data: "RFg5i4,roomNode,washok",         to: "nm.Info" }
    { data: "RFg5i5,roomNode,woonkamer",      to: "nm.Info" }
    { data: "RFg5i6,roomNode,hal vloer",      to: "nm.Info" }
    { data: "RFg5i9,homePower,meterkast",     to: "nm.Info" }
    { data: "RFg5i10,roomNode,hal voor",      to: "nm.Info" }
    { data: "RFg5i11,roomNode,logeerkamer",   to: "nm.Info" }
    { data: "RFg5i12,roomNode,boekenkast L",  to: "nm.Info" }
    { data: "RFg5i13,roomNode,raam halfhoog", to: "nm.Info" }
    { data: "RFg5i14,otRelay,zolderkamer",    to: "nm.Info" }
    { data: "RFg5i15,smaRelay,washok",        to: "nm.Info" }
    { data: "RFg5i18,p1scanner,meterkast",    to: "nm.Info" }
    { data: "RFg5i19,ookRelay,werkkamer",     to: "nm.Info" }
    { data: "RFg5i23,roomNode,gang boven",    to: "nm.Info" }
    { data: "RFg5i24,roomNode,zolderkamer",   to: "nm.Info" }
  ]
  labels: [
    { external: "In", internal: "nm.In" }
    { external: "Out", internal: "nm.Out" }
  ]

# pipeline used for decoding RF12demo data and storing it in the database
circuits.rf12toDatabase =
  gadgets: [
    { name: "st", type: "SketchType" }
    { name: "d1", type: "Dispatcher" }
    { name: "nm", type: "nodesJeeLabs" }
    { name: "d2", type: "Dispatcher" }
    { name: "rd", type: "Readings" }
    { name: "ss", type: "PutReadings" }
    { name: "f2", type: "FanOut" }
    { name: "sr", type: "SplitReadings" }
    { name: "db", type: "LevelDB" }
  ]
  wires: [
    { from: "st.Out", to: "d1.In" }
    { from: "d1.Out", to: "nm.In" }
    { from: "nm.Out", to: "d2.In" }
    { from: "d2.Out", to: "rd.In" }
    { from: "rd.Out", to: "ss.In" }
    { from: "ss.Out", to: "f2.In" }
    { from: "f2.Out:sr", to: "sr.In" }
    { from: "f2.Out:db", to: "db.In" }
    { from: "sr.Out", to: "db.In" }
  ]
  feeds: [
    { data: "./data", to: "db.Name" }
    { data: "Sketch-", to: "d1.Prefix" }
    { data: "Node-", to: "d2.Prefix" }
  ]
  labels: [
    { external: "In", internal: "st.In" }
    { external: "Out", internal: "db.Mods" }
  ]

# serial port test
circuits.serial =
  gadgets: [
    { name: "sp", type: "SerialPort" }
    { name: "ts", type: "TimeStamp" }
    { name: "f1", type: "FanOut" }
    { name: "lg", type: "Logger" }
    { name: "db", type: "rf12toDatabase" }
  ]
  wires: [
    { from: "sp.From", to: "ts.In" }
    { from: "ts.Out", to: "f1.In" }
    { from: "f1.Out:lg", to: "lg.In" }
    { from: "f1.Out:db", to: "db.In" }
  ]
  feeds: [
    { data: "/dev/tty.usbserial-A901ROSN", to: "sp.Port" }
    { data: "./logger", to: "lg.Dir" }
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
  
# simple never-ending demo
circuits.demo =
  gadgets: [
    { name: "c", type: "Clock" }
  ]
  feeds: [
    { data: "1s", to: "c.Rate" }
  ]
  
# pre-load some driver info into the database
circuits.driverFill =
  gadgets: [
    { name: "db", type: "LevelDB" }
  ]
  feeds: [
    { data: "./data", to: "db.Name" }
    { to: "db.In", tag: "/driver/roomNode/temp", \
      data: { name: "Temperature", unit: "°C", scale: 1 } }
    { to: "db.In", tag: "/driver/roomNode/humi", \
      data: { name: "Humidity", unit: "%" } }
    { to: "db.In", tag: "/driver/roomNode/light", \
      data: { name: "Light intensity", unit: "%", factor: 0.392, scale: 0 } }
    { to: "db.In", tag: "/driver/roomNode/moved", \
      data: { name: "Motion", unit: "(0/1)" } }
      
    { to: "db.In", tag: "/driver/smaRelay/yield", \
      data: { name: "PV daily yield", unit: "kWh", scale: 3 } }
    { to: "db.In", tag: "/driver/smaRelay/dcv1", \
      data: { name: "PV level east", unit: "V", scale: 2 } }
    { to: "db.In", tag: "/driver/smaRelay/dcv2", \
      data: { name: "PV level west", unit: "V", scale: 2 } }
    { to: "db.In", tag: "/driver/smaRelay/acw", \
      data: { name: "PV power AC", unit: "W" } }
    { to: "db.In", tag: "/driver/smaRelay/dcw1", \
      data: { name: "PV power east", unit: "W" } }
    { to: "db.In", tag: "/driver/smaRelay/dcw2", \
      data: { name: "PV power west", unit: "W" } }
    { to: "db.In", tag: "/driver/smaRelay/total", \
      data: { name: "PV total", unit: "MWh", scale: 3 } }
      
    { to: "db.In", tag: "/driver/homePower/c1", \
      data: { name: "Counter stove", unit: "kWh", factor: 0.5, scale: 3 } }
    { to: "db.In", tag: "/driver/homePower/p1", \
      data: { name: "Usage stove", unit: "W", scale: 1 } }
    { to: "db.In", tag: "/driver/homePower/c2", \
      data: { name: "Counter solar", unit: "kWh", factor: 0.5, scale: 3 } }
    { to: "db.In", tag: "/driver/homePower/p2", \
      data: { name: "Production solar", unit: "W", scale: 1 } }
    { to: "db.In", tag: "/driver/homePower/c3", \
      data: { name: "Counter house", unit: "kWh", factor: 0.5, scale: 3 } }
    { to: "db.In", tag: "/driver/homePower/p3", \
      data: { name: "Usage house", unit: "W", scale: 1 } }
  ]
  
# pre-load some table info into the database
circuits.tableFill =
  gadgets: [
    { name: "db", type: "LevelDB" }
  ]
  feeds: [
    { data: "./data", to: "db.Name" }

    { to: "db.In", tag: "/table/table", data: { attr: "id attr" } }
    { to: "db.In", tag: "/column/table/id", data: { name: "Ident" } }
    { to: "db.In", tag: "/column/table/attr", data: { name: "Attributes" } }

    { to: "db.In", tag: "/table/column", data: { attr: "id name" } }
    { to: "db.In", tag: "/column/column/id", data: { name: "Ident" } }
    { to: "db.In", tag: "/column/column/name", data: { name: "Name" } }

    { to: "db.In", tag: "/table/driver", data: { attr: "id name unit factor scale" } }
    { to: "db.In", tag: "/column/driver/id", data: { name: "Parameter" } }
    { to: "db.In", tag: "/column/driver/name", data: { name: "Name" } }
    { to: "db.In", tag: "/column/driver/unit", data: { name: "Unit" } }
    { to: "db.In", tag: "/column/driver/factor", data: { name: "Factor" } }
    { to: "db.In", tag: "/column/driver/scale", data: { name: "Scale" } }

    { to: "db.In", tag: "/table/reading", data: { attr: "id loc val ms typ" } }
    { to: "db.In", tag: "/column/reading/id", data: { name: "Ident" } }
    { to: "db.In", tag: "/column/reading/loc", data: { name: "Location" } }
    { to: "db.In", tag: "/column/reading/val", data: { name: "Values" } }
    { to: "db.In", tag: "/column/reading/ms", data: { name: "Timestamp" } }
    { to: "db.In", tag: "/column/reading/typ", data: { name: "Type" } }
  ]

# trial circuit
circuits.try1 =
  gadgets: [
    { name: "db", type: "LevelDB" }
  ]
  feeds: [
    { data: "./data", to: "db.Name" }
    { tag: "<range>", data: "/reading/", to: "db.In" }
  ]

# write configuration to file, but keep a backup of the original, just in case
fs = require 'fs'
try fs.renameSync 'setup.json', 'setup-prev.json'
fs.writeFileSync 'setup.json', JSON.stringify circuits, null, 4
