package dbmodels

import (
	"fmt"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/strongo/validation"
	"strings"
)

// WithSpaceIDs holds SpaceIDs property
type WithSpaceIDs struct {
	SpaceIDs []coretypes.SpaceID `json:"spaceIDs,omitempty" firestore:"spaceIDs,omitempty"`
}

// WithSingleSpaceID returns WithSpaceIDs with single spaceID
func WithSingleSpaceID(spaceID coretypes.SpaceID) WithSpaceIDs {
	return WithSpaceIDs{SpaceIDs: []coretypes.SpaceID{spaceID}}
}

// Validate returns error if not valid
func (v WithSpaceIDs) Validate() error {
	if len(v.SpaceIDs) == 0 {
		return validation.NewErrRecordIsMissingRequiredField("spaceIDs")
	}
	for i, id := range v.SpaceIDs {
		if strings.TrimSpace(string(id)) == "" {
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("spaceIDs[%v]", i), "can not be empty string")
		}
	}
	return nil
}

func (v WithSpaceIDs) JoinSpaceIDs(sep string) string {
	spaceIDs := make([]string, len(v.SpaceIDs))
	for i, s := range v.SpaceIDs {
		spaceIDs[i] = string(s)
	}
	return strings.Join(spaceIDs, sep)
}
