package dbmodels

import (
	"github.com/strongo/validation"
	"strings"
)

// WithTeamID holds TeamID property
type WithTeamID struct {
	TeamID string `json:"teamID" firestore:"teamID"`
}

// Validate returns error if not valid
func (v WithTeamID) Validate() error {
	if strings.TrimSpace(v.TeamID) == "" {
		return validation.NewErrRecordIsMissingRequiredField("teamID")
	}
	return nil
}
