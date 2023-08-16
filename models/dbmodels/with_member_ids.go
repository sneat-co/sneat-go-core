package dbmodels

import (
	"fmt"
	"github.com/strongo/validation"
	"strings"
)

// WithMemberIDs defines a record with a list of member IDs
type WithMemberIDs struct {
	MemberIDs []string `json:"memberIDs,omitempty" firestore:"memberIDs,omitempty"`
}

// Validate returns error as soon as 1st member ContactID is not valid.
func (v WithMemberIDs) Validate() error {
	for i, id := range v.MemberIDs {
		if strings.TrimSpace(id) == "" {
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("memberIDs[%v]", i), "can not be empty string")
		}
	}
	return nil
}

// HasMemberID checks if entity has a specific member ContactID
func (v WithMemberIDs) HasMemberID(id string) bool {
	for _, v := range v.MemberIDs {
		if v == id {
			return true
		}
	}
	return false
}
