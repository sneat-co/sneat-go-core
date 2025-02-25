package dbmodels

import (
	coretype "github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/strongo/validation"
	"strings"
)

// WithSpaceID holds SpaceID property
type WithSpaceID struct {
	SpaceID coretype.SpaceID `json:"spaceID" firestore:"spaceID"`
}

// Validate returns error if not valid
func (v WithSpaceID) Validate() error {
	if strings.TrimSpace(string(v.SpaceID)) == "" {
		return validation.NewErrRecordIsMissingRequiredField("spaceID")
	}
	return nil
}
