package coretypes

import (
	"fmt"
	"strings"
)

// SpaceType is a type of a space, e.g. "private", "family", "company", "space", "club", etc.
type SpaceType string

const (
	// SpaceTypePersonal is a "personal" space type used for a user's personal/home space.
	// (A distinct restricted-visibility "private" type may be introduced later if a
	// real use case arises; it was intentionally not reserved as an unused value.)
	SpaceTypePersonal SpaceType = "personal"

	// SpaceTypeFamily is a "family" space type
	SpaceTypeFamily SpaceType = "family"

	// SpaceTypeGroup is a member-managed group space, such as a Circleus
	// circle. NewSpaceRef constructs its full references; NewWeakSpaceRef
	// deliberately rejects this type because a group reference needs an ID.
	SpaceTypeGroup SpaceType = "group"

	// SpaceTypeCompany is a "company" space type: an organisation whose staff
	// are members and whose public are customers. This is what a school, a
	// gaming venue and a community centre all register as — what the Space is
	// *for* is recorded in SpaceDbo.Modules, not in its type. See sneat-specs
	// decision 0006 (unified space registration).
	SpaceTypeCompany SpaceType = "company"

	// SpaceTypeSpace is a "space" space type.
	//
	// No longer issued for new registrations: decision 0006 has products
	// register a type that describes how membership works, and this one
	// describes nothing. It stays valid so existing records still read, and is
	// deliberately NOT marked Deprecated — the tests that cover legacy records
	// must keep naming it without tripping a linter.
	SpaceTypeSpace SpaceType = "space"

	// SpaceTypeClub is a "club" space type: a members' organisation, where
	// membership *is* the relationship (a sports club's players, guardians,
	// coaches and volunteers are members, not customers).
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

	// SpaceTypeTeam is a child Space of an organisation Space — a club's squad
	// ("Girls U14"), created via create_child_space and never registered as a
	// vendor Space in its own right. It always carries ParentSpaceID; the
	// parent's type says how the organisation's membership works, the team
	// type says this Space is a subdivision of it.
	SpaceTypeTeam SpaceType = "team"
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

// PersonalWeekSpaceRef is a weak space reference for the personal space.
const PersonalWeekSpaceRef = SpaceRef(SpaceTypePersonal)

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
	case SpaceTypeFamily, SpaceTypePersonal:
		return SpaceRef(spaceType)
	default:
		panic(fmt.Sprintf("only 'family' and 'personal' space types are supported for weak space referencing at the moment, got: %s", spaceType))
	}
}

// IsValidSpaceType checks if space has a valid/known type
func IsValidSpaceType(v SpaceType) bool {
	switch v {
	case SpaceTypePersonal, SpaceTypeFamily, SpaceTypeGroup, SpaceTypeCompany, SpaceTypeSpace, SpaceTypeClub, SpaceTypeSystem, SpaceTypeSpot, SpaceTypeTeam:
		return true
	default:
		return false
	}
}

// legacySpaceTypeAliases maps space-type values that were once written to
// records but have since been renamed. Stored data outlives enums: a user
// record from 2025 carries a space brief of type "private" (renamed to
// "personal"), and rejecting it on load bricks every flow that runs through a
// user worker — including registration, whose audience is the least likely to
// have pristine records.
//
// The map is for READ tolerance only: request validation stays strict, so no
// new record is ever written with a legacy value.
var legacySpaceTypeAliases = map[SpaceType]SpaceType{
	"private": SpaceTypePersonal,
}

// CanonicalSpaceType maps a legacy stored space-type value to its current
// name, and returns any other value unchanged. Callers validating STORED data
// should check IsValidSpaceType(CanonicalSpaceType(v)); callers validating
// REQUESTS should keep calling IsValidSpaceType(v) directly, so the legacy
// names stay unwritable.
func CanonicalSpaceType(v SpaceType) SpaceType {
	if canonical, ok := legacySpaceTypeAliases[v]; ok {
		return canonical
	}
	return v
}
