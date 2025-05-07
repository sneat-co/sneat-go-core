package dbo4sharing

import (
	"fmt"
	"github.com/sneat-co/sneat-go-core/sharing/const4sharing"
	"github.com/strongo/strongoapp/with"
	"github.com/strongo/validation"
)

type OfferDbo struct {
	with.CreatedFields
	Permissions []const4sharing.Permission `json:"permissions" firestore:"permissions"`
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
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("permissions[%d]", i), "empty string")
		}
	}
	return nil
}
