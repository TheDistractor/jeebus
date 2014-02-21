// adapted from github,com/codegangsta/gin/lib

package jeebus_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jcw/jeebus"
)

var rpcReply interface{}

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

func mockReply(t *testing.T) func(r interface{}, e error) {
	rpcReply = ""
	return func(r interface{}, e error) {
		if e == nil {
			rpcReply = r
		} else {
			rpcReply = "ERR: " + e.Error()
		}
	}
}

func serveOneRequest(method, urlStr string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	req, _ := http.NewRequest(method, urlStr, nil)
	jeebus.ServeHTTP(response, req)
	return response
}

func contains(list []string, value string) bool {
	for _, x := range list {
		if x == value {
			return true
		}
	}
	return false
}

func wrapArgs(args ...interface{}) []interface{} {
	return args
}

type SpyInfo struct {
	a string
	b []byte
}

type SpyService chan SpyInfo

func (s *SpyService) Handle(topic string, payload []byte) {
	*s <- SpyInfo{topic, payload}
}

func newSpyService() SpyService {
	return make(SpyService, 1)
}
