package extension

import (
	"net/http"
	"testing"
)

func TestHttpHandleFunc(t *testing.T) {
	var f = func(method, path string, handler http.HandlerFunc) {
	}
	f("GET", "/apicore/endpoint", nil)
}
