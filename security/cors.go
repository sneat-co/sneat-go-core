package security

import (
	"github.com/sneat-co/sneat-go/src/core"
	"github.com/strongo/slice"
	"strings"
)

var knownOrigins []string

func init() {
	for _, h := range core.KnownHosts {
		knownOrigins = append(knownOrigins, "https://"+h)
		if h == "localhost" || strings.HasPrefix(h, "localhost:") {
			knownOrigins = append(knownOrigins, "http://"+h)
		}
	}
}

// IsSupportedOrigin verifies if provided origin is allowed
func IsSupportedOrigin(origin string) bool {
	if origin == "" {
		return true
	}
	if slice.Index(knownOrigins, origin) >= 0 {
		return true
	}
	if strings.HasPrefix(origin, "http://localhost:") {
		return true
	}
	return false
}
