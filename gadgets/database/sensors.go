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

// Convert each loosely structured sensor object into a strict map for storage.
func (g *SensorSave) Run() {
	for m := range g.In {
		r := m.(map[string]flow.Message)

		values := r["reading"].(map[string]int)
		asof, ok := r["asof"].(time.Time)
		if !ok {
			asof = time.Now()
		}
		node := r["node"].(map[string]int)
		if node["rssi"] != 0 {
			values["rssi"] = node["rssi"]
		}
		rf12 := r["rf12"].(map[string]int)

		id := fmt.Sprintf("RF12:%d:%d", rf12["group"], node["<node>"])
		data := map[string]interface{}{
			"ms":  asof.UnixNano() / 1000000,
			"val": values,
			"loc": r["location"].(string),
			"typ": r["decoder"].(string),
			"id":  id,
		}
		g.Out.Send(flow.Tag{"/sensor/" + id, data})
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
		data := m.(flow.Tag).Msg.(map[string]interface{})
		for k, v := range data["val"].(map[string]int) {
			key := fmt.Sprintf("reading/%s/%s/%d",
				data["loc"].(string), k, data["ms"].(int64))
			g.Out.Send(flow.Tag{key, v})
		}
	}
}
