package dbmodels

import (
	"github.com/strongo/validation"
)

// PersonPhone holds person's phone number
type PersonPhone struct {
	Type     string `json:"type" firestore:"type"`
	Number   string `json:"number" firestore:"number"`
	Verified bool   `json:"verified" firestore:"verified"`
	Note     string `json:"note,omitempty" firestore:"note,omitempty"`
}

// Validate returns error if not valid
func (v PersonPhone) Validate() error {
	if v.Type == "" {
		return validation.NewErrRecordIsMissingRequiredField("type")
	}
	if v.Number == "" {
		return validation.NewErrRecordIsMissingRequiredField("number")
	}
	return nil
}
