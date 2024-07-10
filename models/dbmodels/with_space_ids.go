package dbmodels

import (
	"fmt"
	"github.com/strongo/validation"
	"strings"
)

// WithSpaceIDs holds SpaceIDs property
type WithSpaceIDs struct {
	SpaceIDs []string `json:"spaceIDs,omitempty" firestore:"spaceIDs,omitempty"`
}

// WithSingleSpaceID returns WithSpaceIDs with single spaceID
func WithSingleSpaceID(spaceID string) WithSpaceIDs {
	return WithSpaceIDs{SpaceIDs: []string{spaceID}}
}

// Validate returns error if not valid
func (v WithSpaceIDs) Validate() error {
	if len(v.SpaceIDs) == 0 {
		return validation.NewErrRecordIsMissingRequiredField("spaceIDs")
	}
	for i, id := range v.SpaceIDs {
		if strings.TrimSpace(id) == "" {
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("spaceIDs[%v]", i), "can not be empty string")
		}
	}
	return nil
}
