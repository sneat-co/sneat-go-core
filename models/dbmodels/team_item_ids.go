package dbmodels

import "errors"
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

func (v TeamItemID) Validate() error {
	s := string(v)
	if s == "" {
		return errors.New("team item ID is empty")
	}
	separatorIndex := strings.IndexByte(s, TeamItemIDSeparatorChar)
	if separatorIndex < 0 {
		return errors.New("team item ID is missing separator char")
	}
	if separatorIndex == 0 {
		return errors.New("team item ID is missing team ID")
	}
	if separatorIndex == len(s)-1 {
		return errors.New("team item ID is missing item ID")
	}
	if strings.IndexByte(s[separatorIndex+1:], TeamItemIDSeparatorChar) >= 0 {
		return errors.New("team item ID has too many separator chars")
	}
	return nil
}
