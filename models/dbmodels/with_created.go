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
	On string `json:"on"`
	By string `json:"by"`
}

func (v *Created) Validate() error {
	var errs []error
	if strings.TrimSpace(v.By) == "" {
		errs = append(errs, validation.NewErrRecordIsMissingRequiredField("by"))
	}
	if strings.TrimSpace(v.On) == "" {
		errs = append(errs, validation.NewErrRecordIsMissingRequiredField("on"))
	}
	if _, err := time.Parse(time.DateOnly, v.On); err != nil {
		return validation.NewErrBadRecordFieldValue("on", err.Error())
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

type WithCreatedField struct {
	Created Created `json:"created"`
}

func (v *WithCreatedField) Validate() error {
	if err := v.Created.Validate(); err != nil {
		return validation.NewErrBadRecordFieldValue("created", err.Error())
	}
	return nil
}

// WithCreated DTO
type WithCreated struct {
	CreatedAt time.Time `json:"createdAt"  firestore:"createdAt"`
	CreatedBy string    `json:"createdBy"  firestore:"createdBy"`
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
