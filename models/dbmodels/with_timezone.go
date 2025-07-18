package dbmodels

import (
	"fmt"
	"github.com/dal-go/dalgo/update"
	"github.com/sneat-co/sneat-go-core"
	"github.com/strongo/validation"
	"strings"
	"time"
)

var _ core.Validatable = (*ByUser)(nil)

// WithTimezone needs to be moved into "with" package
type WithTimezone struct {
	Timezone *Timezone `json:"timezone,omitempty" firestore:"timezone,omitempty"`
}

func (v *WithTimezone) Validate() error {
	return v.Timezone.Validate()
}

func (v *WithTimezone) SetTimezone(loc *time.Location) (updates []update.Update, err error) {
	if loc == nil {
		panic("required argument is nil: WithTimezone.SetTimezone(loc==nil)")
	}
	var offsetMinutes int
	if offsetMinutes, err = getOffsetMinutes(loc, time.Now()); err != nil {
		return
	}
	tzName := loc.String()
	if v.Timezone == nil || v.Timezone.Iana != tzName || v.Timezone.OffsetMinutes != offsetMinutes {
		v.Timezone = &Timezone{
			Iana:          tzName,
			OffsetMinutes: offsetMinutes,
		}
		updates = append(updates, update.ByFieldName("timezone", v.Timezone))
	}
	return
}

func getOffsetMinutes(loc *time.Location, t time.Time) (int, error) {
	_, offsetSeconds := t.In(loc).Zone()
	return offsetSeconds / 60, nil // Convert to minutes
}

// Timezone record
type Timezone struct { // https://www.iana.org/time-zones
	Iana          string `json:"iana,omitempty" firestore:"iana,omitempty"`
	OffsetMinutes int    `json:"offsetMinutes,omitempty" firestore:"offsetMinutes,omitempty"`
}

// Validate validates Timezone record
func (v *Timezone) Validate() error {
	if v == nil {
		return nil
	}
	if v.Iana == "" {
		return validation.NewErrRecordIsMissingRequiredField("iana")
	}
	if !strings.Contains(v.Iana, "/") {
		switch v.Iana {
		case "UTC", "GMT": // OK
		default:
			return validation.NewErrBadRecordFieldValue("iana", "should be UTC or GMT or have / separator")
		}
	} else if slashes := strings.Count(v.Iana, "/"); slashes > 1 {
		return validation.NewErrBadRecordFieldValue("iana", fmt.Sprintf("should be at most 1 '/' separator, got %d", slashes))
	}
	if v.OffsetMinutes%15 != 0 {
		return validation.NewErrBadRecordFieldValue("offsetMinutes",
			fmt.Sprintf("should be divided by 15, got: %d", v.OffsetMinutes))
	}
	return nil
}
