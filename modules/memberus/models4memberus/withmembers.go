package models4memberus

import (
	"errors"
	"fmt"
	"github.com/sneat-co/sneat-go-core/modules/memberus/briefs4memberus"
	"github.com/strongo/slice"
	"github.com/strongo/validation"
)

// WithMembers holds Members property
type WithMembers struct {
	Members []*briefs4memberus.MemberBrief `json:"members,omitempty" firestore:"members,omitempty"`
}

// Validate returns error if not valid
func (v WithMembers) Validate() error {
	for i, m1 := range v.Members {
		if err := m1.Validate(); err != nil {
			return fmt.Errorf("invalid members at index %d: %w", i, err)
		}
		m1.Roles = slice.RemoveInPlace(briefs4memberus.TeamMemberRoleTeamMember, m1.Roles)
		for j, m2 := range v.Members {
			if i != j {
				//if m1.ID == m2.ID {
				//	return errors.New("duplicate members ContactID: " + m1.ID)
				//}
				if m1.UserID != "" && m1.UserID == m2.UserID {
					return errors.New("duplicate members UserID: " + m1.UserID)
				}
				if m1.ShortTitle != "" && m1.ShortTitle == m2.ShortTitle {
					return validation.NewErrBadRecordFieldValue(
						fmt.Sprintf("members[%v].shortTitle", i),
						"duplicate value: "+m1.ShortTitle)
				}
			}
		}
	}
	return nil
}

// GetMemberBriefByUserID returns member's brief by user's ContactID
func (v WithMembers) GetMemberBriefByUserID(userID string) *briefs4memberus.MemberBrief {
	for _, m := range v.Members {
		if m.UserID == userID {
			return m
		}
	}
	return nil
}
