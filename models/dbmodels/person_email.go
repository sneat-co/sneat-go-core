package dbmodels

import (
	"fmt"
	"github.com/strongo/validation"
	"net/mail"
)

// PersonEmail holds person's email
type PersonEmail struct {
	Type     string `json:"type" firestore:"type"`
	Address  string `json:"address" firestore:"address"`
	Verified bool   `json:"verified" firestore:"verified"`
	Note     string `json:"note,omitempty" firestore:"note,omitempty"`
	//
	AuthProvider string `json:"authProvider,omitempty" firestore:"authProvider,omitempty"`
}

func IsKnownAuthProviderID(v string) bool {
	switch v {
	case "password", "google.com":
		return true
	}
	return false
}

// Validate returns error if not valid
func (v PersonEmail) Validate() error {
	switch v.Type {
	case "":
		return validation.NewErrRecordIsMissingRequiredField("type")
	case "primary", "personal", "work":
	// Are known values
	default:
		return validation.NewErrBadRecordFieldValue("type", "unknown value: "+v.Type)
	}
	if v.AuthProvider != "" && !IsKnownAuthProviderID(v.AuthProvider) {
		return validation.NewErrBadRecordFieldValue("authProvider", "unknown value: "+v.AuthProvider)
	}
	if v.Address == "" {
		return validation.NewErrRecordIsMissingRequiredField("address")
	}
	if _, err := mail.ParseAddress(v.Address); err != nil {
		return validation.NewErrBadRecordFieldValue(
			"address",
			fmt.Errorf("invalid email: %v", err).Error(),
		)
	}
	return nil
}
