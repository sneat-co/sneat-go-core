package dbmodels

import (
	"fmt"
	"github.com/strongo/validation"
	"strings"
)

// Address is a postal address
type Address struct {
	CountryID string `json:"countryID" firestore:"countryID"` // ISO 3166-1 alpha-2
	ZipCode   string `json:"zipCode,omitempty" firestore:"zipCode,omitempty"`
	State     string `json:"state,omitempty" firestore:"state,omitempty"`
	City      string `json:"city,omitempty" firestore:"city,omitempty"`
	Lines     string `json:"lines,omitempty" firestore:"lines,omitempty"`
}

// Validate returns error if Address is not valid
func (v *Address) Validate() error {
	if v == nil {
		return nil
	}
	if v.CountryID == "" {
		return validation.NewErrRecordIsMissingRequiredField("countryID")
	}
	if err := ValidateRequiredCountryID("countryID", v.CountryID); err != nil {
		return err
	}
	if strings.TrimSpace(v.ZipCode) != v.ZipCode {
		return validation.NewErrBadRecordFieldValue("zipCode", "should be trimmed")
	}
	lines := strings.TrimSpace(v.Lines)

	validateLen := func(field, v string, max int) error {
		if len(v) > max {
			return validation.NewErrBadRecordFieldValue(field, fmt.Sprintf("should be less then %d characters", max))
		}
		return nil
	}
	if err := validateLen("city", v.City, 85); err != nil {
		return err
	}
	if err := validateLen("state", v.State, 30); err != nil {
		return err
	}
	if len(lines) != len(v.Lines) {
		return validation.NewErrBadRecordFieldValue("lines", "should be trimmed")
	}
	if err := validateLen("lines", v.Lines, 1000); err != nil {
		return err
	}
	return nil
}
