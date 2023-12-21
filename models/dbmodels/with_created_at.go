package dbmodels

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/validation"
	"time"
)

type WithCreatedAt struct {
	CreatedAt string `json:"createdAt"  firestore:"createdAt"`
}

// GetCreatedAt returns value of CreatedAt field as time.Time parsed with RFC3339Nano layout
func (v *WithCreatedAt) GetCreatedAt() (time.Time, error) {
	return time.Parse(time.RFC3339Nano, v.CreatedAt)
}

// SetCreatedAt sets CreatedAt field formatted with RFC3339Nano layout
func (v *WithCreatedAt) SetCreatedAt(t time.Time) {
	v.CreatedAt = t.Format(time.RFC3339Nano)
}

func (v *WithCreatedAt) UpdatesCreatedOn() []dal.Update {
	return []dal.Update{
		{Field: "createdOn", Value: v.CreatedAt},
	}
}

func (v *WithCreatedAt) Validate() error {
	if v.CreatedAt == "" {
		return validation.NewErrRecordIsMissingRequiredField("createdAt")
	}
	if _, err := time.Parse(time.DateOnly, v.CreatedAt); err != nil {
		return validation.NewErrBadRecordFieldValue("createdAt", err.Error())
	}
	return nil
}
