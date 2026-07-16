package coretypes

import (
	"fmt"
	"strings"
)

// SpaceType is a type of a space, e.g. "private", "family", "company", "space", "club", etc.
type SpaceType string

const (
	// SpaceTypePrivate is a "private" space type
	SpaceTypePrivate SpaceType = "private"

	// SpaceTypeFamily is a "family" space type
	SpaceTypeFamily SpaceType = "family"

	SpaceTypeGroup = "group"

	// SpaceTypeCompany is a "company" space type
	SpaceTypeCompany SpaceType = "company"

	// SpaceTypeSpace is a "space" space type
	SpaceTypeSpace SpaceType = "space"

	// SpaceTypeClub is a "club" space type
	SpaceTypeClub SpaceType = "club"

	// SpaceTypeSystem is a platform-owned "system" space type for shared,
	// cross-user records that are not tied to per-user membership.
	SpaceTypeSystem SpaceType = "system"

	// SpaceTypeSpot is a venue-scoped space that owns a ToGethered Spot's shared
	// records (happenings, spot-days, and — after migration — all spot-scoped
	// ToGethered records).  Membership is empty in MVP: followers are
	// subscriptions, not members; moderators may become members in a later
	// phase.
	SpaceTypeSpot SpaceType = "spot"
)

// SpotSpaceIDPrefix is the reserved prefix used by SpotSpaceID to construct
// deterministic, human-readable space IDs for ToGethered Spots.
// The "~" separator is chosen because it cannot appear in randomly-generated
// Firestore document IDs (which use base-62 characters) and is not used by
// any other SpaceID separator in the codebase (SpaceRefSeparator="!",
// SpaceItemIDSeparator="_"), so it is guaranteed not to collide.
const SpotSpaceIDPrefix = "spot~"

// SpotSpaceID returns the deterministic SpaceID for the ToGethered Spot with
// the given spotID.  The ID has the form "spot~<spotID>", e.g. "spot~acme-gym".
// The "spot~" prefix is reserved and must not be used for any other purpose.
func SpotSpaceID(spotID string) SpaceID {
	return SpaceID(SpotSpaceIDPrefix + spotID)
}

const FamilyWeekSpaceRef = SpaceRef(SpaceTypeFamily)
const PrivateWeekSpaceRef = SpaceRef(SpaceTypePrivate)

type SpaceRef string

func (v SpaceRef) SpaceType() SpaceType {
	if i := strings.Index(string(v), SpaceRefSeparator); i >= 0 {
		return SpaceType(v[:i])
	}
	if IsValidSpaceType(SpaceType(v)) {
		return SpaceType(v)
	}
	return ""
}

// SpaceID returns space userID from the space reference
func (v SpaceRef) SpaceID() SpaceID {
	if i := strings.Index(string(v), SpaceRefSeparator); i >= 0 {
		return SpaceID(v[i+1:])
	}
	if !IsValidSpaceType(SpaceType(v)) {
		return SpaceID(v)
	}
	return ""
}

// UrlPath returns a URL path for the space reference
func (v SpaceRef) UrlPath() string {
	return fmt.Sprintf("%s/%s", v.SpaceType(), v.SpaceID())
}

const SpaceRefSeparator = "!"

// NewSpaceRef creates a new SpaceRef
func NewSpaceRef(spaceType SpaceType, spaceID SpaceID) SpaceRef {
	if !IsValidSpaceType(spaceType) {
		panic(fmt.Errorf("invalid space type: %v", spaceType))
	}
	if spaceID == "" {
		panic("spaceID is an empty string")
	}
	return SpaceRef(string(spaceType) + SpaceRefSeparator + string(spaceID))
}

// NewWeakSpaceRef creates a new weak SpaceRef, e.g. only with space type, no space userID
func NewWeakSpaceRef(spaceType SpaceType) SpaceRef {
	switch spaceType {
	case SpaceTypeFamily, SpaceTypePrivate:
		return SpaceRef(spaceType)
	default:
		panic(fmt.Sprintf("only 'family' and 'private' space types are supported for weak space referencing at the moment, got: %s", spaceType))
	}
}

// IsValidSpaceType checks if space has a valid/known type
func IsValidSpaceType(v SpaceType) bool {
	switch v {
	case SpaceTypeFamily, SpaceTypePrivate, SpaceTypeGroup, SpaceTypeCompany, SpaceTypeSpace, SpaceTypeClub, SpaceTypeSystem, SpaceTypeSpot:
		return true
	default:
		return false
	}
}
