package dbmodels

import (
	"fmt"
	"github.com/strongo/slice"
	"github.com/strongo/validation"
	"strings"
)

// WithOptionalCountryID defines a record with a Country ContactID
type WithOptionalCountryID struct {
	CountryID string `json:"countryID,omitempty" firestore:"countryID"` // Intentionally do NOT omitempty for Firestore
}

func (v WithOptionalCountryID) Validate() error {
	if err := ValidateOptionalCountryID("countryID", v.CountryID); err != nil {
		return validation.NewErrBadRequestFieldValue("WithOptionalCountryID", err.Error())
	}
	return nil
}

// WithRequiredCountryID defines a record with a Country ContactID
type WithRequiredCountryID struct {
	CountryID string `json:"countryID" firestore:"countryID"`
}

func (v WithRequiredCountryID) Validate() error {
	if err := ValidateRequiredCountryID("countryID", v.CountryID); err != nil {
		return err
	}
	return nil
}

func ValidateRequiredCountryID(field, value string) error {
	return ValidateCountryID(field, value, true)
}

func ValidateCountryID(field, value string, isRequired bool) error {
	if isRequired && strings.TrimSpace(value) == "" {
		return validation.NewErrRecordIsMissingRequiredField(field)
	}
	return ValidateOptionalCountryID(field, value)
}

func ValidateOptionalCountryID(field, value string) error {
	if value == "" {
		return nil
	}
	if strings.TrimSpace(value) != value {
		return validation.NewErrBadRecordFieldValue(field, "leading or closing spaces")
	}
	var countryID string
	if i := strings.Index(value, ":"); i < 0 {
		countryID = value
	} else {
		panic("Unclear case usage, consider removal or decoupling")
		//countryID = value[:i]
		//if i == len(value)-1 {
		//	return validation.NewErrBadRecordFieldValue(field, "empty suffix")
		//}
	}
	if l := len(countryID); l != 2 {
		return validation.NewErrBadRecordFieldValue(field, fmt.Sprintf("countryID expected to be 2 charactes, got %v", l))
	}
	if strings.ToUpper(countryID) != countryID {
		return validation.NewErrBadRecordFieldValue(field, fmt.Sprintf("should be in upper case: %v", countryID))
	}
	if slice.Index(knownCountryIDs, countryID) < 0 {
		return validation.NewErrBadRecordFieldValue(field, fmt.Sprintf("unknown countryID: %v", countryID))
	}
	return nil
}

const UnknownCountryID = "--"

var knownCountryIDs = []string{
	UnknownCountryID,
	"AR",
	"AU",
	"BR",
	"CA",
	"CL",
	"CN",
	"CO",
	"DE",
	"EE",
	"EG",
	"ES",
	"ES",
	"FR",
	"GB",
	"IE",
	"IN",
	"IT",
	"JP",
	"KE",
	"LV",
	"LT",
	"MX",
	"NG",
	"NZ",
	"PE",
	"RU",
	"UA",
	"UK",
	"US",
	"VE",
	"ZA",
	"ZA",
}
