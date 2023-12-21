package with

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/validation"
	"time"
)

type CreatedAtField struct {
	CreatedAt string `json:"createdAt" dalgo:"createdAt" firestore:"createdAt"`
}

// GetCreatedAt returns value of CreatedAtField field as time.Time parsed with RFC3339Nano layout
func (v *CreatedAtField) GetCreatedAt() (time.Time, error) {
	return time.Parse(time.RFC3339Nano, v.CreatedAt)
}

// SetCreatedAt sets CreatedAtField field formatted with RFC3339Nano layout
func (v *CreatedAtField) SetCreatedAt(t time.Time) {
	v.CreatedAt = t.Format(time.RFC3339Nano)
}

func (v *CreatedAtField) UpdatesCreatedOn() []dal.Update {
	return []dal.Update{
		{Field: "createdOn", Value: v.CreatedAt},
	}
}

func (v *CreatedAtField) Validate() error {
	if v.CreatedAt == "" {
		return validation.NewErrRecordIsMissingRequiredField("createdAt")
	}
	if _, err := time.Parse(time.DateOnly, v.CreatedAt); err != nil {
		return validation.NewErrBadRecordFieldValue("createdAt", err.Error())
	}
	return nil
}
