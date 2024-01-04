package security

import (
	"slices"
	"strings"
)

var knownOrigins []string

func init() {
	for _, host := range knownHosts {
		addKnownOrigins(host)
	}
}

func addKnownOrigins(host string) {
	knownOrigins = append(knownOrigins, "https://"+host)
	if host == "localhost" || strings.HasPrefix(host, "localhost:") {
		knownOrigins = append(knownOrigins, "http://"+host)
	}
}

// IsSupportedOrigin verifies if provided origin is allowed
func IsSupportedOrigin(origin string) bool {
	if origin == "" {
		return true
	}
	if slices.Contains(knownOrigins, origin) {
		return true
	}
	if strings.HasPrefix(origin, "http://localhost:") {
		return true
	}
	return false
}
