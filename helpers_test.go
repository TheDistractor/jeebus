// adapted from github,com/codegangsta/gin/lib

package jeebus_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/jcw/jeebus"
)

var (
	onceMessaging sync.Once
)

func init() {
	log.SetFlags(log.Ltime)
}

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

func onceStartMessaging(t *testing.T) {
	onceMessaging.Do(func() {
		err := jeebus.StartMessaging(nil)
		expect(t, err, nil)
	})
}

func serveOneRequest(method, urlStr string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	req, _ := http.NewRequest(method, urlStr, nil)
	jeebus.ServeHTTP(response, req)
	return response
}

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
