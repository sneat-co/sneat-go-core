package httpserver

import (
	"fmt"
	"github.com/sneat-co/sneat-go/src/core/security"
	"net/http"
)

// AccessControlAllowOrigin verifies HTTP header "Origin"
func AccessControlAllowOrigin(w http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if !security.IsSupportedOrigin(origin) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = fmt.Fprintf(w, "Unsupported origin: %v", origin)
		return false
	}
	if header := w.Header(); header.Get("Access-Control-Allow-Origin") == "" {
		header.Add("Access-Control-Allow-Origin", origin)
	}
	return true
}
