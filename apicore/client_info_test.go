package apicore

import (
	"net/http/httptest"
	"testing"
)

func TestGetRemoteClientInfo(t *testing.T) {
	r := httptest.NewRequest("GET", "http://localhost/", nil)
	_ = GetRemoteClientInfo(r)
}
