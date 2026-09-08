package security

import (
	"slices"
	"strings"
)

var knownHosts = []string{
	"localhost:4200",
	"local-app.sneat.ws",
}

var knownHostSuffixes []string

// IsKnownHost checks if it is a known app
func IsKnownHost(host string) bool {
	return slices.Contains(knownHosts, host)
}

func AddKnownHosts(hosts ...string) {
	for _, host := range hosts {
		if !slices.Contains(knownHosts, host) {
			knownHosts = append(knownHosts, host)
			addKnownOrigins(host)
		}
	}
}

// AddKnownHostSuffixes allows HTTPS origins on subdomains of an explicitly
// Sneat-owned suffix. It does not allow the suffix itself, ports, HTTP, or
// lookalike suffixes. Use AddKnownHosts for individual customer-owned hosts.
func AddKnownHostSuffixes(suffixes ...string) {
	for _, suffix := range suffixes {
		suffix = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(suffix)), ".")
		suffix = strings.TrimPrefix(strings.TrimPrefix(suffix, "*."), ".")
		if !isValidHostname(suffix) {
			panic("invalid known host suffix: " + suffix)
		}
		if !slices.Contains(knownHostSuffixes, suffix) {
			knownHostSuffixes = append(knownHostSuffixes, suffix)
		}
	}
}
