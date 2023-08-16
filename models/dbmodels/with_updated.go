package dbmodels

import (
	"errors"
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/validation"
	"strings"
	"time"
)

// WithUpdated DTO
type WithUpdated struct {
	UpdatedAt time.Time `json:"updatedAt,omitempty"  firestore:"updatedAt,omitempty"`
	UpdatedBy string    `json:"updatedBy,omitempty"  firestore:"updatedBy,omitempty"`
}

// UpdatesWhenUpdated populates update instructions for DAL when a record has been updated
func (v *WithUpdated) UpdatesWhenUpdated() []dal.Update {
	return []dal.Update{
		{Field: "updatedAt", Value: v.UpdatedAt},
		{Field: "updatedBy", Value: v.UpdatedBy},
	}
}

// Validate returns error if not valid
func (v *WithUpdated) Validate() error {
	var errs []error
	if v.UpdatedAt.IsZero() {
		errs = append(errs, validation.NewErrRecordIsMissingRequiredField("updatedAt"))
	}
	if strings.TrimSpace(v.UpdatedBy) == "" {
		errs = append(errs, validation.NewErrRecordIsMissingRequiredField("updatedBy"))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
