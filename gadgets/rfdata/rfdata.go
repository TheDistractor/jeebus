// Driver and decoders for RF12/RF69 packet data.
package rfdata

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jcw/flow"
)

func init() {
	flow.Registry["Sketch-RF12demo"] = func() flow.Circuitry { return &RF12demo{} }
	flow.Registry["NodeMap"] = func() flow.Circuitry { return &NodeMap{} }
	flow.Registry["Readings"] = func() flow.Circuitry { return &Readings{} }
	flow.Registry["PutReadings"] = func() flow.Circuitry { return &PutReadings{} }
	flow.Registry["SplitReadings"] = func() flow.Circuitry { return &SplitReadings{} }
}

// RF12demo parses config and OK lines coming from the RF12demo sketch.
// Registers as "Sketch-RF12demo".
type RF12demo struct {
	flow.Gadget
	In  flow.Input  // serial input, as tex lines
	Out flow.Output // <node> map, followed by []byte packet
	Oob flow.Output // the same, for out-of-band packets
	Rej flow.Output // unrecognised strings
}

// Start converting lines into binary packets.
func (w *RF12demo) Run() {
	for m := range w.In {
		s, ok := m.(string)
		if !ok {
			w.Out.Send(m) // not a string
			continue
		}
		if strings.HasPrefix(s, "[RF12demo.") {
			w.Out.Send(parseConfigLine(s))
		}
		if strings.HasPrefix(s, "OK ") {
			data, rssi := convertToBytes(s)
			info := map[string]int{"<node>": int(data[0] & 0x1F)}
			if rssi != 0 {
				info["rssi"] = rssi
			}
			if data[0]&0xA0 == 0xA0 {
				w.Oob.Send(info)
				w.Oob.Send(data)
			} else {
				w.Out.Send(info)
				w.Out.Send(data)
			}
		} else {
			w.Rej.Send(m)
		}
	}
}

// Parse lines of the form "[RF12demo.12] _ i31* g5 @ 868 MHz c1 q1"
var re = regexp.MustCompile(`\.(\d+)] . i(\d+)\*? g(\d+) @ (\d+) MHz`)

func parseConfigLine(s string) map[string]int {
	m := re.FindStringSubmatch(s)
	v, _ := strconv.Atoi(m[1])
	i, _ := strconv.Atoi(m[2])
	g, _ := strconv.Atoi(m[3])
	b, _ := strconv.Atoi(m[4])
	return map[string]int{"<RF12demo>": v, "band": b, "group": g, "id": i}
}

func convertToBytes(s string) ([]byte, int) {
	s = strings.TrimSpace(s[3:])
	var rssi int

	// convert a line of decimal byte values to a byte buffer
	var buf bytes.Buffer
	for _, v := range strings.Split(s, " ") {
		if strings.HasPrefix(v, "(") {
			rssi, _ = strconv.Atoi(v[1 : len(v)-1])
		} else {
			n, _ := strconv.Atoi(v)
			buf.WriteByte(byte(n))
		}
	}
	return buf.Bytes(), rssi
}

// Lookup the group/node information to determine what decoder to use.
// Registers as "NodeMap".
type NodeMap struct {
	flow.Gadget
	Info flow.Input
	In   flow.Input
	Out  flow.Output
}

// Start looking up node ID's in the node map.
func (w *NodeMap) Run() {
	nodeMap := map[string]string{}
	locations := map[string]string{}
	for m := range w.Info {
		f := strings.Split(m.(string), ",")
		nodeMap[f[0]] = f[1]
		if len(f) > 2 {
			locations[f[0]] = f[2]
		}
	}

	var group int
	for m := range w.In {
		w.Out.Send(m)

		if data, ok := m.(map[string]int); ok {
			switch {
			case data["<RF12demo>"] > 0:
				group = data["group"]
			case data["<node>"] > 0:
				key := fmt.Sprintf("RFg%di%d", group, data["<node>"])
				if loc, ok := locations[key]; ok {
					w.Out.Send(flow.Tag{"<location>", loc})
				}
				if tag, ok := nodeMap[key]; ok {
					w.Out.Send(flow.Tag{"<dispatch>", tag})
				}
			}
		}
	}
}

// Re-combine decoded readings into single objects.
type Readings struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
}

// Start listening and combining readings produced by various decoders.
func (g *Readings) Run() {
	state := map[string]flow.Message{}
	other := []flow.Message{}
	for m := range g.In {
		switch v := m.(type) {
		case time.Time:
			state["asof"] = v
		case map[string]int:
			switch {
			case v["<RF12demo>"] > 0:
				state["rf12"] = v
			case v["<node>"] > 0:
				state["node"] = v
				delete(state, "location")
			case v["<reading>"] > 0:
				delete(v, "<reading>")
				state["reading"] = v
				if len(other) > 0 {
					state["other"] = other
					other = []flow.Message{}
				}
				g.Out.Send(state)
			default:
				other = append(other, v)
			}
		case flow.Tag:
			switch v.Tag {
			case "<dispatched>":
				state["decoder"] = v.Msg
			case "<location>":
				state["location"] = v.Msg
			default:
				other = append(other, v)
			}
		default:
			other = append(other, v)
		}
	}
}

// Save readings in database.
type PutReadings struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
}

// Convert each loosely structured reading object into a strict map for storage.
func (g *PutReadings) Run() {
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
		location, _ := r["location"].(string)
		decoder, _ := r["decoder"].(string)

		id := fmt.Sprintf("RF12:%d:%d", rf12["group"], node["<node>"])
		data := map[string]interface{}{
			"ms":  asof.UnixNano() / 1000000,
			"val": values,
			"loc": location,
			"typ": decoder,
			"id":  id,
		}
		g.Out.Send(flow.Tag{"/reading/" + id, data})
	}
}

// Split reading data into individual values.
type SplitReadings struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
}

// Split combined readings into separate sensor values, for separate storage.
func (g *SplitReadings) Run() {
	for m := range g.In {
		data := m.(flow.Tag).Msg.(map[string]interface{})
		for k, v := range data["val"].(map[string]int) {
			key := fmt.Sprintf("sensor/%s/%s/%d",
				data["loc"].(string), k, data["ms"].(int64))
			g.Out.Send(flow.Tag{key, v})
		}
	}
}
