package dbmodels

import (
	"fmt"
	"github.com/dal-go/dalgo/update"
	"github.com/sneat-co/sneat-go-core"
	"github.com/strongo/validation"
	"regexp"
)

var _ core.Validatable = (*ByUser)(nil)

// WithTimezone needs to be moved into "with" package
type WithTimezone struct {
	Timezone *Timezone `json:"timezone,omitempty" firestore:"timezone,omitempty"`
}

func (v *WithTimezone) Validate() error {
	return v.Timezone.Validate()
}

func (v *WithTimezone) SetTimezone(iana string, utcOffset string) (updates []update.Update) {
	if v.Timezone == nil || v.Timezone.Iana != iana || v.Timezone.UtcOffset != utcOffset {
		v.Timezone = &Timezone{
			Iana:      iana,
			UtcOffset: utcOffset,
		}
		updates = append(updates, update.ByFieldName("timezone", v.Timezone))
	}
	return
}

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
