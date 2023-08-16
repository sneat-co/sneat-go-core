package dbmodels

import (
	"github.com/sneat-co/sneat-go-core"
	"github.com/strongo/validation"
)

var _ core.Validatable = (*ByUser)(nil)

// ByUser record
type ByUser struct {
	UID   string `json:"uid" firestore:"uid"`
	Title string `json:"title,omitempty" firestore:"title,omitempty"`
}

// Validate validates ByUser record
func (v *ByUser) Validate() error {
	if v.UID == "" {
		return validation.NewErrRecordIsMissingRequiredField("uid")
	}
	return nil
}
