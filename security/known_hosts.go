package security

import (
	"slices"
)

var knownHosts = []string{
	"localhost:4200",
	"local.sneat.ws",
}

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
