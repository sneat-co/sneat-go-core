package models4memberus

import (
	"github.com/sneat-co/sneat-go-core/modules/teamus/core4teamus"
	"github.com/strongo/validation"
	"strings"
)

// TeamBrief holds brief info about a team a member is a part of
type TeamBrief struct {
	ID    string               `json:"id" firestore:"id"`
	Type  core4teamus.TeamType `json:"type" firestore:"type"`
	Title string               `json:"title" firestore:"title"`
}

// Validate returns error if not valid
func (v TeamBrief) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return validation.NewErrRecordIsMissingRequiredField("id")
	}
	if !core4teamus.IsValidTeamType(v.Type) {
		return validation.NewErrRecordIsMissingRequiredField("type")
	}
	if strings.TrimSpace(v.Title) == "" {
		return validation.NewErrRecordIsMissingRequiredField("title")
	}
	return nil
}
