// Package dbo4acl is the per-record access-control primitive: an ACL maps a
// userID to the set of permissions that user has been granted on a record.
//
// It is the grant core shared by record-level authorization and by the richer
// sharing model (sharing/dbo4sharing.To embeds ACL and adds cross-space sharing).
//
// specscore: decisions/0002-reserved-extension-space-ids
package dbo4acl

import (
	"github.com/sneat-co/sneat-go-core/acl/const4acl"
	"github.com/strongo/strongoapp/with"
	"github.com/strongo/validation"
)

// Permissions is the set of permissions granted to a single user, keyed by
// permission ID, each carrying who granted it and when.
type Permissions map[string]with.CreatedFields

// Validate returns an error if any granted permission is malformed.
func (v Permissions) Validate() error {
	for id, created := range v {
		switch id {
		case
			const4acl.PermittedToView,
			const4acl.PermittedToEdit,
			const4acl.PermittedToRsvp:
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
