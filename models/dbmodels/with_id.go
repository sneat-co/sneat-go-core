package dbmodels

import "github.com/strongo/validation"

// DtoWithID is a DTO with ID. Mostly intended to add ID to a brief DTO
type DtoWithID[T interface{ Validate() error }] struct {
	ID   string `json:"id" firestore:"id"`
	Data T      `json:"data" firestore:"data"`
}

func (v *DtoWithID[T]) Validate() error {
	if v == nil {
		return nil
	}
	if v.ID == "" {
		return validation.NewErrRecordIsMissingRequiredField("id")
	}
	if err := v.Data.Validate(); err != nil {
		return validation.NewErrBadRecordFieldValue("data", err.Error())
	}
	return nil
}
