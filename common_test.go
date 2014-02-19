package jeebus_test

import (
	"net/http"
	"testing"
)

func TestCommonViaServer(t *testing.T) {
	response := serveOneRequest("GET", "/common/")
	expect(t, response.Code, http.StatusOK)
	refute(t, response.Body.Len(), 0)
}
