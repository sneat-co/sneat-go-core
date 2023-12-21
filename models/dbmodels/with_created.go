package dbmodels

import (
	"errors"
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/validation"
	"strings"
	"time"
)

// Created is intended to be used only in WithCreatedField. For root level use WithCreated instead.
type Created struct {
	At string `json:"at" dalgo:"at" firestore:"at"`
	By string `json:"by" dalgo:"at" firestore:"by"`
}

// Validate returns error if not valid
func (v *Created) Validate() error {
	var errs []error
	if strings.TrimSpace(v.At) == "" {
		errs = append(errs, validation.NewErrRecordIsMissingRequiredField("at"))
	}
	if strings.TrimSpace(v.By) == "" {
		errs = append(errs, validation.NewErrRecordIsMissingRequiredField("by"))
	}
	if _, err := time.Parse(time.DateOnly, v.At); err != nil {
		return validation.NewErrBadRecordFieldValue("on", err.Error())
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// WithCreatedField adds a Created field to a data model
type WithCreatedField struct {
	Created Created `json:"created" firestore:"created"`
}

func (v *WithCreatedField) Validate() error {
	if err := v.Created.Validate(); err != nil {
		return validation.NewErrBadRecordFieldValue("created", err.Error())
	}
	return nil
}

// WithCreated adds CreatedAt and CreatedBy fields to a data model
type WithCreated struct {
	WithCreatedAt
	WithCreatedBy
}

// UpdatesWithCreated populates update instructions for DAL when a record has been created
func (v *WithCreated) UpdatesWithCreated() []dal.Update {
	return append(v.WithCreatedAt.UpdatesCreatedOn(), v.WithCreatedBy.UpdatesCreatedBy()...)
}

// Validate returns error if not valid
func (v *WithCreated) Validate() error {
	var errs []error
	if err := v.WithCreatedAt.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := v.WithCreatedBy.Validate(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
