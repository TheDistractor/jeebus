// Driver and decoders for RF12/RF69 packet data.
package rfdata

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jcw/flow"
)

func init() {
	flow.Registry["Sketch-RF12demo"] = func() flow.Circuitry { return &RF12demo{} }
	flow.Registry["NodeMap"] = func() flow.Circuitry { return &NodeMap{} }
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
	if m, ok := <-w.In; ok {
		// config := parseConfigLine(m.(string))
		config := parseConfigLine("[RF12demo.12] _ i31* g5 @ 868 MHz c1 q1")
		w.Out.Send(config)
		for m = range w.In {
			if s, ok := m.(string); ok {
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
			} else {
				w.Out.Send(m) // not a string
			}
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
		f := strings.SplitN(m.(string), " ", 3)
		nodeMap[f[0]] = f[1]
		if len(f) > 2 {
			locations[f[0]] = f[2]
		}
	}

	var group int
	for m := range w.In {
		if data, ok := m.(map[string]int); ok {
			w.Out.Send(m)

			switch {
			case data["<RF12demo>"] > 0:
				group = data["group"]
			case data["<node>"] > 0:
				key := fmt.Sprintf("RFg%di%d", group, data["<node>"])
				if loc, ok := locations[key]; ok {
					w.Out.Send(flow.Tag{"<location>", loc})
				}
				if tag, ok := nodeMap[key]; ok {
					w.Out.Send(flow.Tag{"<dispatch>", "Node-" + tag})
				}
			}
			continue
		}

		w.Out.Send(m)
	}
}