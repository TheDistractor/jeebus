package jeebus_test

import (
	"testing"

	"github.com/jcw/jeebus"
)

func TestToolsApp(t *testing.T) {
	refute(t, jeebus.NewApp("", ""), nil)
}
