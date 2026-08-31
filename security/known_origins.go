package security

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"slices"
	"strconv"
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
	if IsLocalhostHost(host) {
		knownOrigins = append(knownOrigins, "http://"+host)
	}
}

var ErrBadOrigin = errors.New("bad origin")

// VerifyOrigin verifies if provided origin is allowed
func VerifyOrigin(origin string) error {
	if origin == "" {
		return nil
	}
	if isLocalhostOrigin(origin) {
		return nil
	}
	if slices.Contains(knownOrigins, origin) {
		return nil
	}
	return fmt.Errorf("%w: %s: known origins: %s", ErrBadOrigin, origin, strings.Join(knownOrigins, ", "))
}

// IsLocalhostHost reports whether host is the reserved localhost name or one
// of its subdomains. host may include a port, as it does in http.Request.Host.
func IsLocalhostHost(host string) bool {
	if strings.Contains(host, ":") {
		hostname, port, err := net.SplitHostPort(host)
		if err != nil {
			return false
		}
		portNumber, err := strconv.Atoi(port)
		if err != nil || portNumber < 1 || portNumber > 65535 {
			return false
		}
		host = hostname
	}
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if !isValidHostname(host) {
		return false
	}
	return host == "localhost" || strings.HasSuffix(host, ".localhost")
}

func isValidHostname(host string) bool {
	if host == "" || len(host) > 253 {
		return false
	}
	for _, label := range strings.Split(host, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, char := range label {
			if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' {
				return false
			}
		}
	}
	return true
}

func isLocalhostOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Opaque != "" {
		return false
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return false
	}
	if parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || strings.HasSuffix(parsed.Host, ":") {
		return false
	}
	return IsLocalhostHost(parsed.Host)
}
