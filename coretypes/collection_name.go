package coretypes

import (
	"errors"
	"strings"
)

// ValidateCollectionName requires one nonempty logical collection component.
// Collection names are serialized as-is by storage adapters, unlike escaped
// record IDs. A collection must never encode a nested path; embedded items use
// the explicit, owner-resolved ItemRef.SubPath contract instead. This does not
// trim, decode, or otherwise change a name, or impose provider-specific naming
// rules on characters that are not path delimiters.
func ValidateCollectionName(name string) error {
	if name == "" {
		return errors.New("collection is required")
	}
	if strings.ContainsRune(name, '/') {
		return errors.New("collection must be a single namespace component")
	}
	return nil
}
