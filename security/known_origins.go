package security

import (
	"errors"
	"fmt"
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

var ErrBadOrigin = errors.New("bad origin")

// VerifyOrigin verifies if provided origin is allowed
func VerifyOrigin(origin string) error {
	if origin == "" {
		return nil
	}
	if strings.HasPrefix(origin, "http://localhost:") {
		return nil
	}
	if slices.Contains(knownOrigins, origin) {
		return nil
	}
	return fmt.Errorf("%w: %s: known origins: %s", ErrBadOrigin, origin, strings.Join(knownOrigins, ", "))
}
