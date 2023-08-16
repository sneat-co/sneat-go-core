package dbmodels

import "strings"

// TeamItemIDSeparator is a separator character between team ID and item ID
const TeamItemIDSeparatorChar = '_'
const TeamItemIDSeparator = "_"

// NewTeamItemID returns team item ID as a concatenation of team ID and item ID
func NewTeamItemID(teamID, id string) TeamItemID {
	return TeamItemID(teamID + TeamItemIDSeparator + id)
}

type TeamItemID string

func (v TeamItemID) TeamID() string {
	return string(v[:strings.IndexByte(string(v), TeamItemIDSeparatorChar)])
}

func (v TeamItemID) ItemID() string {
	return string(v[strings.IndexByte(string(v), TeamItemIDSeparatorChar)+1:])
}
