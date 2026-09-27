package apicoretest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTestEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// AuthRequired = false
	TestEndpoint(t, func(w http.ResponseWriter, r *http.Request) {}, AssertOptions{AuthRequired: false}, req)

	// AuthRequired = true
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}
	TestEndpoint(t, handler, AssertOptions{AuthRequired: true}, req)
}
