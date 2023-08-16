package dbmodels

import (
	"github.com/strongo/validation"
	"strings"
)

var _ UserIDGetter = (*WithUserID)(nil)

// WithUserID defines a record with a user ContactID
type WithUserID struct {
	UserID string `json:"userID,omitempty" firestore:"userID,omitempty"`
}

func (v WithUserID) GetUserID() string {
	return v.UserID
}

// Validate returns error if user ContactID is not valid
func (v WithUserID) Validate() error {
	if v.UserID != "" {
		if strings.TrimSpace(v.UserID) != v.UserID {
			return validation.NewErrBadRecordFieldValue("userID", "leading or closing spaces")
		}
	}
	return nil
}

type UserIDGetter interface {
	GetUserID() string
}
