package dbmodels

import (
	"fmt"
	"github.com/sneat-co/sneat-go/src/core"
	"github.com/strongo/validation"
	"regexp"
)

var _ core.Validatable = (*ByUser)(nil)

// Timezone record
type Timezone struct { // https://www.iana.org/time-zones
	Iana      string `json:"iana,omitempty" firestore:"iana,omitempty"`
	UtcOffset string `json:"utcOffset,omitempty" firestore:"utcOffset,omitempty"`
}

var reTimezoneOffset = regexp.MustCompile(`[+-][01]\d+:[0-5][05]`)

// Validate validates Timezone record
func (v *Timezone) Validate() error {
	if v == nil {
		return nil
	}
	if v.Iana == "" {
		return validation.NewErrRecordIsMissingRequiredField("iana")
	}
	if v.UtcOffset != "" && !reTimezoneOffset.MatchString(v.UtcOffset) {
		return validation.NewErrBadRecordFieldValue("utcOffset",
			fmt.Sprintf("does not match '±HH:MM' pattern, got: [%v]", v.UtcOffset))
	}
	return nil
}
