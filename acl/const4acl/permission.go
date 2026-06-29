// Package const4acl defines the permission vocabulary for per-record access
// control (ACL). These capabilities are referenced by an ACL's per-user grants.
//
// specscore: decisions/0002-reserved-extension-space-ids
package const4acl

// Permission is a capability that may be granted to a user on a record.
type Permission = string

const (
	// PermittedToView grants read access to a record.
	PermittedToView Permission = "view"
	// PermittedToEdit grants write access to a record.
	PermittedToEdit Permission = "edit"
	// PermittedToRsvp grants the ability to RSVP to a record (e.g. an event).
	PermittedToRsvp Permission = "rsvp"
)
