package jeebus_test

import (
	"net/http"
	"testing"
)

func TestFilesViaServer(t *testing.T) {
	response := serveOneRequest("GET", "/files/")
	expect(t, response.Code, http.StatusOK)
	refute(t, response.Body.Len(), 0)
}
