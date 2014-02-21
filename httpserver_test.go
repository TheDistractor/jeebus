package jeebus_test

import (
	"net/http"
	"testing"
)

func TestAppServer(t *testing.T) {
	response := serveOneRequest("GET", "/")
	expect(t, response.Code, http.StatusOK)
	refute(t, response.Body.Len(), 0)
}

func TestAnyPathServer(t *testing.T) {
	response := serveOneRequest("GET", "/blah") // still served, same as "/"
	expect(t, response.Code, http.StatusOK)
	refute(t, response.Body.Len(), 0)
}

func TestAnyFileServer(t *testing.T) {
	response := serveOneRequest("GET", "/blah.blah") // this one contains a "."
	expect(t, response.Code, http.StatusNotFound)
}
