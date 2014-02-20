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

// FIXME: yuck, copy of same code in helpers_test.go
type SpyInfo struct {
	a string
	b interface{}
}

type SpyService chan SpyInfo

func (s *SpyService) Handle(topic string, payload interface{}) {
	*s <- SpyInfo{topic, payload}
}

func newSpyService() SpyService {
	return make(SpyService, 1)
}

func TestJeeBusImport(t *testing.T) {
	jeebus.Check(nil)
}
