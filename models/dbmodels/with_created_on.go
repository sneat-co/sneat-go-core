package dbmodels

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/validation"
	"time"
)

type WithCreatedOn struct {
	CreatedOn string `json:"createdOn"  firestore:"createdOn"`
}

func (v *WithCreatedOn) UpdatesCreatedOn() []dal.Update {
	return []dal.Update{
		{Field: "createdOn", Value: v.CreatedOn},
	}
}

func (v *WithCreatedOn) Validate() error {
	if v.CreatedOn == "" {
		return validation.NewErrRecordIsMissingRequiredField("createdOn")
	}
	if _, err := time.Parse(time.DateOnly, v.CreatedOn); err != nil {
		return validation.NewErrBadRecordFieldValue("createdOn", err.Error())
	}
	return nil
}
