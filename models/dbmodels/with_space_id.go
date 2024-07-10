package dbmodels

import (
	"github.com/strongo/validation"
	"strings"
)

// WithSpaceID holds SpaceID property
type WithSpaceID struct {
	SpaceID string `json:"spaceID" firestore:"spaceID"`
}

// Validate returns error if not valid
func (v WithSpaceID) Validate() error {
	if strings.TrimSpace(v.SpaceID) == "" {
		return validation.NewErrRecordIsMissingRequiredField("spaceID")
	}
	return nil
}
