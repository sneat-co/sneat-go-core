package briefs4contactus

import (
	"github.com/sneat-co/sneat-go-core"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/models/dbprofile"
	"github.com/strongo/validation"
	"strings"
)

type ContactBrief struct {
	dbmodels.WithUserID
	Type       ContactType     `json:"type" firestore:"type"` // "person", "company", "location"
	Gender     dbmodels.Gender `json:"gender,omitempty" firestore:"gender,omitempty"`
	Name       *dbmodels.Name  `json:"name,omitempty" firestore:"name,omitempty"`
	Title      string          `json:"title,omitempty" firestore:"title,omitempty"`
	ShortTitle string          `json:"shortTitle,omitempty" firestore:"shortTitle,omitempty"` // Not supposed to be used in models4contactus.ContactDto
	ParentID   string          `json:"parentID" firestore:"parentID"`                         // Intentionally not adding `omitempty` so we can search root contacts only

	// Number of active invites to join a team
	InvitesCount int `json:"activeInvitesCount,omitempty" firestore:"activeInvitesCount,omitempty"`

	// AgeGroup is deprecated?
	AgeGroup string `json:"ageGroup,omitempty" firestore:"ageGroup,omitempty"` // TODO: Add validation
	// Avatar holds a photo of a member
	Avatar                         *dbprofile.Avatar `json:"avatar,omitempty" firestore:"avatar,omitempty"`
	dbmodels.WithOptionalRelatedAs                   // This is used in `RelatedContacts` field of `ContactDto`
	dbmodels.WithOptionalCountryID
	dbmodels.WithRoles
}

// GetUserID returns UserID field value
func (v *ContactBrief) GetUserID() string {
	return v.UserID
}

// Equal returns true if 2 instances are equal
func (v *ContactBrief) Equal(v2 *ContactBrief) bool {
	return v.Type == v2.Type &&
		v.WithUserID == v2.WithUserID &&
		v.Gender == v2.Gender &&
		v.WithOptionalCountryID == v2.WithOptionalCountryID &&
		v.Name.Equal(v2.Name) &&
		v.WithOptionalRelatedAs.Equal(v2.WithOptionalRelatedAs) &&
		v.Avatar.Equal(v2.Avatar)
}

// Validate returns error if not valid
func (v *ContactBrief) Validate() error {
	if err := ValidateContactType(v.Type); err != nil {
		return err
	}
	if err := dbmodels.ValidateGender(v.Gender, false); err != nil {
		return err
	}
	if strings.TrimSpace(v.Title) == "" && v.Name == nil {
		return validation.NewErrRecordIsMissingRequiredField("name|title")
	} else if err := v.Name.Validate(); err != nil {
		return validation.NewErrBadRecordFieldValue("name", err.Error())
	}
	if v.UserID != "" {
		if !core.IsAlphanumericOrUnderscore(v.UserID) {
			return validation.NewErrBadRecordFieldValue("userID", "is not alphanumeric: "+v.UserID)
		}
	}
	switch v.Type {
	case ContactTypeLocation:
		if v.ParentID == "" {
			return validation.NewErrRecordIsMissingRequiredField("parentID")
		}
	}
	if err := v.WithOptionalCountryID.Validate(); err != nil {
		return err
	}
	if err := v.WithRoles.Validate(); err != nil {
		return err
	}
	if err := v.WithUserID.Validate(); err != nil {
		return err
	}
	return nil
}

// GetTitle return full name of a person
func (v *ContactBrief) GetTitle() string {
	if v.Title != "" {
		return v.Title
	}
	if v.Name.Full != "" {
		return v.Name.Full
	}
	if v.Name.First != "" && v.Name.Last != "" && v.Name.Middle != "" {
		return v.Name.First + " " + v.Name.Middle + " " + v.Name.Full
	}
	if v.Name.First != "" && v.Name.Last != "" {
		return v.Name.First + " " + v.Name.Full
	}
	if v.Name.First != "" {
		return v.Name.First
	}
	if v.Name.Last != "" {
		return v.Name.Last
	}
	if v.Name.Middle != "" {
		return v.Name.Middle
	}
	return ""
}
