package core

import "github.com/strongo/slice"

var KnownHosts = []string{
	"sneat.app",
	"logistus.app",
	"sneatapp.web.app",
	"sneatapp.firebaseapp.com",
	"sneat-team.web.app",
	"sneat-team.firebaseapp.com",
	"dailyscrum.app",
	"localhost:4200",
}

// IsKnownHost checks if it is a known app
func IsKnownHost(app string) bool {
	return slice.Index[string](KnownHosts, app) >= 0
}
