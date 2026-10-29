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

func TestGetRemoteClientInfoOrigin(t *testing.T) {
	for _, tt := range []struct{ origin, referer, want string }{
		{origin: "https://datatug.app", want: "https://datatug.app"},
		{referer: "https://sneat.app/signup?x=1", want: "https://sneat.app"},
		{origin: "https://datatug.app", referer: "https://sneat.app/", want: "https://datatug.app"},
		{origin: "null", referer: "https://sneat.app/a", want: "https://sneat.app"},
		{origin: "not a url", want: ""},
		{want: ""},
	} {
		r := httptest.NewRequest("POST", "http://api.sneat.ws/", nil)
		if tt.origin != "" {
			r.Header.Set("Origin", tt.origin)
		}
		if tt.referer != "" {
			r.Header.Set("Referer", tt.referer)
		}
		if got := GetRemoteClientInfo(r).Origin; got != tt.want {
			t.Errorf("origin=%q referer=%q: got %q, want %q", tt.origin, tt.referer, got, tt.want)
		}
	}
}
