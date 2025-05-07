package dbo4sharing

import (
	"fmt"
	"github.com/sneat-co/sneat-go-core/sharing"
	"github.com/strongo/strongoapp/with"
	"github.com/strongo/validation"
)

type Shared struct {
	ID          string              `json:"id" firestore:"id"` // This an ID of the shared item in the receiver space
	Permissions sharing.Permissions `json:"permissions" firestore:"permissions"`
}

type To struct {
	Spaces map[string]Shared              `json:"spaces" firestore:"spaces"`
	Users  map[string]sharing.Permissions `json:"users" firestore:"users"`
}

type OfferDbo struct {
	with.CreatedFields
	Permissions []sharing.Permission `json:"permissions" firestore:"permissions"`
}

func (v OfferDbo) Validate() error {
	if err := v.CreatedFields.Validate(); err != nil {
		return err
	}
	if len(v.Permissions) == 0 {
		return validation.NewErrRecordIsMissingRequiredField("permissions")
	}
	for i, p := range v.Permissions {
		if p == "" {
			return validation.NewErrBadRecordFieldValue(fmt.Sprint("permissions[%d]", i), "empty string")
		}
	}
	return nil
}
