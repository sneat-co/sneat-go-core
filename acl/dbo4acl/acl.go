package dbo4acl

import (
	"github.com/sneat-co/sneat-go-core/acl/const4acl"
	"github.com/strongo/validation"
)

// ACL is the per-record access-control list: it records which users have been
// granted which permissions on the record that carries it. It is a storage of
// grants, not an authorization scope — authorization is always evaluated per
// record against this list (see Decision 0002).
type ACL struct {
	// Users maps a userID to the set of permissions granted to that user.
	Users map[string]Permissions `json:"users" firestore:"users"`
}

// Validate returns an error if the ACL is malformed.
func (v ACL) Validate() error {
	for id, permissions := range v.Users {
		if id == "" {
			return validation.NewErrBadRecordFieldValue("users", "has empty user ID")
		}
		if err := permissions.Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue("users["+id+"]", err.Error())
		}
	}
	return nil
}

// UserCan reports whether userID has been granted permission p by this record's
// ACL. An empty userID, an absent user, or an absent grant all deny: there is no
// implicit/blanket permission — the decision comes solely from the record's own
// grants.
func (v ACL) UserCan(userID string, p const4acl.Permission) bool {
	if userID == "" {
		return false
	}
	permissions, ok := v.Users[userID]
	if !ok {
		return false
	}
	_, ok = permissions[p]
	return ok
}
