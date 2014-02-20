package opt_test

import (
	"sync"
	"testing"

	"github.com/jcw/jeebus"
)

var onceMessaging sync.Once

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (%T) - Got %v (%T)", b, b, a, a)
	}
}

func refute(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Errorf("Did not expect %v (%T) - Got %v (%T)", b, b, a, a)
	}
}

func TestJeeBusImport(t *testing.T) {
	jeebus.Check(nil)
}

func onceStartMessaging(t *testing.T) {
	onceMessaging.Do(func() {
		err := jeebus.StartMessaging(nil)
		expect(t, err, nil)
	})
}
