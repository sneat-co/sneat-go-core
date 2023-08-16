package dbmodels

import (
	"fmt"
	"github.com/strongo/validation"
	"strings"
)

// WithTeamIDs holds TeamIDs property
type WithTeamIDs struct {
	TeamIDs []string `json:"teamIDs,omitempty" firestore:"teamIDs,omitempty"`
}

// WithSingleTeamID returns WithTeamIDs with single teamID
func WithSingleTeamID(teamID string) WithTeamIDs {
	return WithTeamIDs{TeamIDs: []string{teamID}}
}

// Validate returns error if not valid
func (v WithTeamIDs) Validate() error {
	if len(v.TeamIDs) == 0 {
		return validation.NewErrRecordIsMissingRequiredField("teamIDs")
	}
	for i, id := range v.TeamIDs {
		if strings.TrimSpace(id) == "" {
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("teamIDs[%v]", i), "can not be empty string")
		}
	}
	return nil
}
