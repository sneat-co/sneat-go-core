package core

import "github.com/strongo/slice"

var KnownHosts = []string{
	"localhost:4200",
}

// IsKnownHost checks if it is a known app
func IsKnownHost(app string) bool {
	return slice.Index[string](KnownHosts, app) >= 0
}
