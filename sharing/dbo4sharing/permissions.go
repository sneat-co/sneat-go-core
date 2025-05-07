package dbo4sharing

import (
	"github.com/sneat-co/sneat-go-core/sharing/const4sharing"
	"github.com/strongo/strongoapp/with"
	"github.com/strongo/validation"
)

type Permissions map[string]with.CreatedFields

func (v Permissions) Validate() error {
	for id, created := range v {
		switch id {
		case
			const4sharing.PermittedToView,
			const4sharing.PermittedToEdit,
			const4sharing.PermittedToRsvp:
			// ok
		case "":
			return validation.NewErrBadRecordFieldValue("permissions", "has empty permission ID")
		}
		if err := created.Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue("permissions", err.Error())
		}
	}
	return nil
}
