package dbmodels

import (
	"errors"
	"github.com/strongo/strongoapp/with"
	"github.com/strongo/validation"
	"strings"
	"time"
)

// Modified DTO
type Modified struct {
	By string    `json:"by,omitempty"  firestore:"by,omitempty"`
	At time.Time `json:"at,omitempty"  firestore:"at,omitempty"`
}

// Validate returns error if not valid
func (v Modified) Validate() error {
	if strings.TrimSpace(v.By) == "" {
		return validation.NewErrRecordIsMissingRequiredField("by")
	}
	if v.At.IsZero() {
		return validation.NewErrRecordIsMissingRequiredField("at")
	}
	return nil
}

// WithModified DTO
type WithModified struct {
	with.CreatedFields
	with.UpdatedFields
	with.DeletedFields
}

func NewWithModified(at time.Time, by string) WithModified {
	var withCreated with.CreatedFields
	withCreated.SetCreatedAt(at)
	withCreated.SetCreatedBy(by)
	return WithModified{
		CreatedFields: withCreated,
		UpdatedFields: with.UpdatedFields{
			UpdatedAt: at,
			UpdatedBy: by,
		},
	}
}

// Validate returns error if not valid
func (v *WithModified) Validate() error {
	var errs []error
	if err := v.CreatedFields.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := v.UpdatedFields.Validate(); err != nil {
		errs = append(errs, validation.NewErrBadRecordFieldValue("updated", err.Error()))
	}
	if err := v.DeletedFields.Validate(); err != nil {
		errs = append(errs, validation.NewErrBadRecordFieldValue("deleted", err.Error()))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// MarkAsUpdated marks record as updated
func (v *WithModified) MarkAsUpdated(uid string) {
	v.UpdatedBy = uid
	v.UpdatedAt = time.Now()
}
