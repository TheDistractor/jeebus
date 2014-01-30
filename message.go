package jeebus

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Message represent a payload over MQTT for a specified topic.
type Message struct {
	T   string                     // topic
	P   json.RawMessage            // payload
	obj map[string]json.RawMessage // decoded payload object fields
}

// String returns a short string representation of a Message.
func (m *Message) String() string {
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
	return fmt.Sprintf("«M:%s,%s»", m.T, strings.Map(f, msg))
}

// unpack the JSON payload into a map, this fails if payload is not an object.
func (m *Message) unpack(key string, v interface{}) {
	if m.obj == nil && len(m.P) > 0 {
		err := json.Unmarshal(m.P, &m.obj)
		check(err)
	}
	json.Unmarshal(m.obj[key], &v)
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
