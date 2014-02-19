package jeebus_test

import (
	"net/http"
	"testing"
)

func TestNotFound(t *testing.T) {
	response := serveOneRequest("GET", "/blah")
	expect(t, response.Code, http.StatusNotFound)
}

func TestAppServer(t *testing.T) {
	response := serveOneRequest("GET", "/")
	expect(t, response.Code, http.StatusOK)
	refute(t, response.Body.Len(), 0)
}
