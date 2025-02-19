package dbmodels

import (
	"errors"
	"github.com/dal-go/dalgo/update"
	"github.com/sneat-co/sneat-go-core"
	"github.com/strongo/strongoapp/with"
	"github.com/strongo/validation"
	"time"
)

// Versioned defines an interface for versioned record
type Versioned interface {
	core.Validatable

	// IncreaseVersion returns new record version increased by 1. It also should update UpdatedAt and UpdatedBy fields.
	IncreaseVersion(updatedAt time.Time, updatedBy string) int
}

type WithVersion struct {
	Version int `json:"v" firestore:"v"`
}

func (v *WithVersion) Validate() error {
	if v.Version < 1 {
		return validation.NewErrBadRecordFieldValue("v", "must be >= 1")
	}
	return nil
}

func (v *WithVersion) IncreaseVersion() {
	v.Version++
}

func (v *WithVersion) GetUpdates() []update.Update {
	return []update.Update{update.ByFieldName("v", v.Version)}
}

type WithUpdatedAndVersion struct {
	with.UpdatedFields
	WithVersion
}

func (v *WithUpdatedAndVersion) IncreaseVersion(updatedAt time.Time, updatedBy string) int {
	v.WithVersion.IncreaseVersion()
	v.UpdatedAt = updatedAt
	v.UpdatedBy = updatedBy
	return v.Version
}

func (v *WithUpdatedAndVersion) Validate() error {
	var errs []error
	if err := v.WithVersion.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := v.UpdatedFields.Validate(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (v *WithUpdatedAndVersion) GetUpdates() []update.Update {
	return append(
		v.WithVersion.GetUpdates(),
		v.UpdatedFields.UpdatesWhenUpdatedFieldsChanged()...,
	)
}
