package httpserver

import (
	"fmt"
	"github.com/sneat-co/sneat-go-core/security"
	"net/http"
)

// AccessControlAllowOrigin verifies HTTP header "Origin"
func AccessControlAllowOrigin(w http.ResponseWriter, r *http.Request) bool {
	if r == nil {
		panic("request is nil")
	}
	if w == nil {
		panic("response writer is nil")
	}
	origin := r.Header.Get("Origin")
	if err := security.VerifyOrigin(origin); err != nil {
		w.WriteHeader(http.StatusForbidden)
		_, _ = fmt.Fprintf(w, "Unsupported origin: %v", err)
		return false
	}
	if header := w.Header(); header.Get("Access-Control-Allow-Origin") == "" {
		header.Add("Access-Control-Allow-Origin", origin)
	}
	return true
}
