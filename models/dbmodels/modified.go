package dbmodels

import (
	"errors"
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
	WithCreated
	WithUpdated
	WithDeleted
}

func NewWithModified(at time.Time, by string) WithModified {
	return WithModified{
		WithCreated: WithCreated{
			CreatedAt: at,
			CreatedBy: by,
		},
		WithUpdated: WithUpdated{
			UpdatedAt: time.Now(),
			UpdatedBy: by,
		},
	}
}

// Validate returns error if not valid
func (v *WithModified) Validate() error {
	var errs []error
	if err := v.WithCreated.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := v.WithUpdated.Validate(); err != nil {
		errs = append(errs, validation.NewErrBadRecordFieldValue("updated", err.Error()))
	}
	if err := v.WithDeleted.Validate(); err != nil {
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
