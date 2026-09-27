package apicore

import (
	"net/http/httptest"
	"testing"
)

func TestGetRemoteClientInfo(t *testing.T) {
	r := httptest.NewRequest("GET", "http://localhost/", nil)
	_ = GetRemoteClientInfo(r)

	r2 := httptest.NewRequest("GET", "http://localhost/", nil)
	r2.Header.Set("CF-Connecting-IP", "1.2.3.4")
	r2.Header.Set("X-Forwarded-For", "1.2.3.4")
	info := GetRemoteClientInfo(r2)
	if info.ForwardedFor != "" {
		t.Fatalf("expected empty ForwardedFor, got: %s", info.ForwardedFor)
	}
}
