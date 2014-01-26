package jeebus

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Message represent a payload over MQTT for a specified topic.
type Message struct {
	T   string                      // topic
	P   json.RawMessage             // payload
	R   bool                        // retain
	obj map[string]*json.RawMessage // decoded payload object fields
}

// String returns a short string representation of a Message.
func (m *Message) String() string {
	// display the retain flag only if set
	retain := ""
	if m.R {
		retain = ",R"
	}
	// insert an ellipsis if the payload data is too long
	// note that all numbers, booleans, and nulls will pass through as is
	msg := string(m.P)
	if len(msg) > 20 {
		msg = msg[:18] + "…"
		switch msg[0] {
		case '{':
			msg += "}"
		case '[':
			msg += "]"
		default:
			msg += msg[:1] // only double quotes, really
		}
	}
	// replace the most common non-printable characters by a dot
	f := func(r rune) rune {
		if r < ' ' {
			r = '.'
		}
		return r
	}
	return fmt.Sprintf("«M:%s,%s%s»", m.T, strings.Map(f, msg), retain)
}

func (m *Message) useMap() {
	if m.obj == nil {
		m.obj = make(map[string]*json.RawMessage)
		if len(m.P) > 0 {
			err := json.Unmarshal(m.P, &m.obj)
			check(err)
		}
	}
}

// unpack the JSON payload into a map, this fails if payload is not an object.
func (m *Message) unpack(key string, v interface{}) {
	m.useMap()
	if p, ok := m.obj[key]; ok {
		json.Unmarshal(*p, &v)
	}
}

// Get extracts a given object attribute as string, or "" if absent.
func (m *Message) Get(key string) (v string) {
	m.unpack(key, &v)
	return
}

// GetBool extracts a given object attribute as bool, or false if absent.
func (m *Message) GetBool(key string) (v bool) {
	m.unpack(key, &v)
	return
}

// GetInt extracts a given object attribute as int, or 0 if absent.
func (m *Message) GetInt(key string) int {
	return int(m.GetFloat64(key))
}

// GetInt64 extracts a given object attribute as 64-bit int, or 0 if absent.
func (m *Message) GetInt64(key string) int64 {
	return int64(m.GetFloat64(key))
}

// GetFloat64 extracts a given object attribute as float, or 0 if absent.
func (m *Message) GetFloat64(key string) (v float64) {
	m.unpack(key, &v)
	return
}

// Set allows setting keys with arbitrary values, for publishing later.
func (m *Message) Set(key string, value interface{}) {
	newVal, err := json.Marshal(value)
	check(err)
	m.useMap()
	// FIXME yuck!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	var x json.RawMessage = newVal // TODO still struggling with casts in Go ...
	m.obj[key] = &x
}

// Publish the current message payload to the given topic.
func (m *Message) Publish(topic string) {
	if m.obj != nil {
		msg, err := json.Marshal(m.obj)
		check(err)
		Publish(topic, msg)
	} else {
		Publish(topic, m.P)
	}
}
