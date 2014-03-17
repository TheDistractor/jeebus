package database

import (
	"fmt"
	"time"

	"github.com/jcw/flow"
)

func init() {
	flow.Registry["SensorSave"] = func() flow.Circuitry { return &SensorSave{} }
	flow.Registry["SplitReadings"] = func() flow.Circuitry { return &SplitReadings{} }
}

// Save readings in database.
type SensorSave struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
}

// The data structure used for saving readings in the database.
type SensorData struct {
	Millis   int64          // milliseconds, works well with JavaScript
	Values   map[string]int // integer readings for one or more parameters
	Location string         // the location of the sensor
	Decoder  string         // the name of the decoder used
	Node     string         // the identity of the node
}

// Convert each loosely structured sensor object into a struct for storage.
func (g *SensorSave) Run() {
	for m := range g.In {
		r := m.(map[string]flow.Message)

		values := r["reading"].(map[string]int)
		asof, ok := r["asof"].(time.Time)
		if !ok {
			asof = time.Now()
		}
		millis := asof.UnixNano() / 1000000
		node := r["node"].(map[string]int)
		if node["rssi"] != 0 {
			values["rssi"] = node["rssi"]
		}
		location := r["location"].(string)
		rf12 := r["rf12"].(map[string]int)

		key := fmt.Sprintf("/sensor/%s/%d", location, millis)
		data := SensorData{
			Millis:   millis,
			Values:   values,
			Location: location,
			Decoder:  r["decoder"].(string),
			Node:     fmt.Sprintf("RF12:%d:%d", rf12["group"], node["<node>"]),
		}
		g.Out.Send(flow.Tag{key, data})
	}
}

// Split sensor data into individual values.
type SplitReadings struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
}

// Split combined measurements into individual readings, for separate storage.
func (g *SplitReadings) Run() {
	for m := range g.In {
		data := m.(flow.Tag).Msg.(SensorData)
		for k, v := range data.Values {
			key := fmt.Sprintf("/reading/%s/%s/%d",
				data.Location, k, data.Millis)
			g.Out.Send(flow.Tag{key, v})
		}
	}
}
