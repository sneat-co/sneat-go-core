package models4memberus

import "github.com/strongo/random"

// RandomMemberID creates a random ContactID for a new member
var RandomMemberID = func() string {
	return random.ID(2)
}
