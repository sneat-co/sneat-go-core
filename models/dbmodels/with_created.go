package dbmodels

import (
	"errors"
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/validation"
	"strings"
	"time"
)

// WithCreated DTO
type WithCreated struct {
	CreatedAt time.Time `json:"createdAt,omitempty"  firestore:"createdAt,omitempty"`
	CreatedBy string    `json:"createdBy,omitempty"  firestore:"createdBy,omitempty"`
}

// UpdatesWhenCreated populates update instructions for DAL when a record has been created
func (v *WithCreated) UpdatesWhenCreated() []dal.Update {
	return []dal.Update{
		{Field: "createdAt", Value: v.CreatedAt},
		{Field: "createdBy", Value: v.CreatedBy},
	}
}

// Validate returns error if not valid
func (v *WithCreated) Validate() error {
	var errs []error
	if v.CreatedAt.IsZero() {
		errs = append(errs, validation.NewErrRecordIsMissingRequiredField("createdAt"))
	}
	if strings.TrimSpace(v.CreatedBy) == "" {
		errs = append(errs, validation.NewErrRecordIsMissingRequiredField("createdBy"))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
