package dbo4sharing

import (
	"github.com/sneat-co/sneat-go-core/acl/dbo4acl"
	"github.com/strongo/validation"
)

// To describes how a record is shared. The embedded ACL holds the per-user
// grants — the access-control core (dbo4acl.ACL) reused for per-record
// authorization — and Spaces additionally shares the record into other spaces.
// Sharing is a superset of an ACL: it is the ACL plus cross-space sharing.
//
// The ACL is embedded anonymously so its `users` field stays at the same path
// (e.g. `sharedTo.users`) — the wire format is unchanged.
type To struct {
	dbo4acl.ACL
	Spaces map[string]Shared `json:"spaces" firestore:"spaces"`
}

func (v To) Validate() error {
	if err := v.ACL.Validate(); err != nil {
		return err
	}
	for id, shared := range v.Spaces {
		if err := shared.Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue("spaces["+id+"]", err.Error())
		}
	}
	return nil
}

type Shared struct {
	ID          string              `json:"id" firestore:"id"` // This an ID of the shared item in the receiver space
	Permissions dbo4acl.Permissions `json:"permissions" firestore:"permissions"`
}

func (v Shared) Validate() error {
	if v.ID == "" {
		return validation.NewErrRecordIsMissingRequiredField("id")
	}
	if len(v.Permissions) == 0 {
		return validation.NewErrRecordIsMissingRequiredField("permissions")
	}

	if err := v.Permissions.Validate(); err != nil {
		return validation.NewErrBadRecordFieldValue("permissions", err.Error())
	}
	return nil
}
