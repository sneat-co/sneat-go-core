package dbmodels

import (
	"github.com/strongo/validation"
	"strings"
)

type Gender = string

const (
	GenderMale            Gender = "male"
	GenderFemale          Gender = "female"
	GenderUnknown         Gender = "unknown"
	GenderNonbinary       Gender = "nonbinary"
	GenderTypeUndisclosed Gender = "undisclosed"
	GenderTypeOther       Gender = "other"
)

// Genders defines known gender
var Genders = []Gender{
	GenderMale,
	GenderFemale,
	GenderNonbinary,
	GenderUnknown,
	GenderTypeOther,
	GenderTypeUndisclosed,
}

// IsKnownGenderOrEmpty returns error if bad value
func IsKnownGenderOrEmpty(v string) bool {
	if v == "" {
		return true
	}
	return IsKnownGender(v)
}

// IsKnownGender returns false if bad value
func IsKnownGender(v string) bool {
	for _, g := range Genders {
		if g == v {
			return true
		}
	}
	return false
}

// ValidateGender checks gender
func ValidateGender(v string, required bool) error {
	if required && strings.TrimSpace(v) == "" {
		return validation.NewErrRecordIsMissingRequiredField("gender")
	}
	if !IsKnownGenderOrEmpty(v) {
		return validation.NewErrBadRecordFieldValue("gender", "unknown value: "+v)
	}
	return nil
}
