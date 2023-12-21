package dbmodels

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/validation"
	"strings"
)

type WithCreatedBy struct {
	CreatedBy string `json:"createdBy"  firestore:"createdBy"`
}

// SetCreatedBy sets CreatedBy field
func (v *WithCreatedBy) SetCreatedBy(createBy string) {
	v.CreatedBy = createBy
}

// GetCreatedBy returns CreatedBy
func (v *WithCreatedBy) GetCreatedBy() string {
	return v.CreatedBy
}

func (v *WithCreatedBy) UpdatesCreatedBy() []dal.Update {
	return []dal.Update{
		{Field: "createdBy", Value: v.CreatedBy},
	}
}

func (v *WithCreatedBy) Validate() error {
	if strings.TrimSpace(v.CreatedBy) == "" {
		return validation.NewErrRecordIsMissingRequiredField("createdBy")
	}
	return nil
}
