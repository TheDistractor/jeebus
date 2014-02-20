// adapted from github,com/codegangsta/gin/lib

package jeebus_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jcw/jeebus"
)

func init() {
	log.SetFlags(log.Ltime)
}

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %T: %v - got %T: %v", b, b, a, a)
	}
}

func refute(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Errorf("Did not expect %T: %v", a, a)
	}
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
