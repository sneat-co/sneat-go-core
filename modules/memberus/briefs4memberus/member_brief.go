package briefs4memberus

import (
	"github.com/sneat-co/sneat-go-core/modules/contactus/briefs4contactus"
)

// GetFullMemberID returns full member ContactID
func GetFullMemberID(teamID, memberID string) string {
	if teamID == "" {
		panic("teamID is required parameter")
	}
	if memberID == "" {
		panic("memberID is required parameter")
	}
	return teamID + ":" + memberID
}

type MemberBrief = briefs4contactus.ContactBrief

// IsUniqueShortTitle checks if a given value is an unique member title
func IsUniqueShortTitle(v string, contacts map[string]*briefs4contactus.ContactBrief, role string) bool {
	for _, c := range contacts {
		if c.ShortTitle == v && (role == "" || c.HasRole(role)) {
			return false
		}
	}
	return true
}
