package coretypes

import (
	"fmt"
	"strings"

	"github.com/dal-go/record"
	core "github.com/sneat-co/sneat-go-core"
)

// spacesCollection is the storage collection name for Space records.
const spacesCollection = "spaces"

// spaceModulesCollection is the storage collection name for a space's
// (or, when spaceID is empty, the spaceless system namespace's) extension
// records. The value matches
// sneat-core-modules/core/coremodels.ExtCollection ("ext"); it is kept as a
// local literal rather than imported to avoid re-creating the
// sneat-core-modules <-> sneat-go-core module import cycle this promotion
// exists to break.
const spaceModulesCollection = "ext"

const maxSpaceIDLength = 30

// ValidateSpaceID applies the canonical constraints used by Space keys without
// panicking. Request DTOs must call this before passing untrusted IDs to
// NewSpaceKey.
func ValidateSpaceID(spaceID SpaceID) error {
	if spaceID == "" {
		return fmt.Errorf("spaceID is empty")
	}
	if length := len(spaceID); length > maxSpaceIDLength {
		return fmt.Errorf("spaceID is %d characters long, maximum is %d", length, maxSpaceIDLength)
	}
	if core.IsAlphanumericOrUnderscore(string(spaceID)) {
		return nil
	}
	// Reserved entity-scoped space IDs carry a '~' separator by design:
	// SpotSpaceID builds "spot~{spotID}" (SpaceTypeSpot venue spaces). Accept
	// the documented prefix with an alphanumeric suffix; everything else stays
	// rejected.
	suffix, ok := strings.CutPrefix(string(spaceID), SpotSpaceIDPrefix)
	if !ok || suffix == "" || !core.IsAlphanumericOrUnderscore(suffix) {
		return fmt.Errorf("spaceID contains unsupported characters or uppercase letters: %q", spaceID)
	}
	return nil
}

// NewSpaceKey create new doc ref
func NewSpaceKey(spaceID SpaceID) *record.Key {
	if err := ValidateSpaceID(spaceID); err != nil {
		panic(err.Error())
	}
	return record.NewKeyWithID(spacesCollection, string(spaceID))
}

// NewSpaceModuleKey builds the storage key for an extension's per-space
// record.
//
// specscore: decisions/0002-reserved-extension-space-ids
// Spaceless key-builder: when the space is empty, build /ext/{ext-id}/... and
// do NOT call NewSpaceKey (which panics on empty). See sneat-specs Decision
// 0002:
// https://github.com/sneat-co/sneat-specs/blob/main/spec/decisions/0002-reserved-extension-space-ids.md
func NewSpaceModuleKey(spaceID SpaceID, moduleID ExtID) *record.Key {
	if spaceID == "" {
		// Spaceless system namespace: global/system extension records live at
		// /ext/{ext-id}/... with no /spaces/{space-id} prefix (sneat-specs
		// Decision 0002). NewSpaceKey is intentionally NOT called here.
		return record.NewKeyWithID(spaceModulesCollection, moduleID)
	}
	spaceKey := NewSpaceKey(spaceID)
	return record.NewKeyWithParentAndID(spaceKey, spaceModulesCollection, moduleID)
}
