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
)

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
	case SpaceTypeFamily, SpaceTypePrivate, SpaceTypeGroup, SpaceTypeCompany, SpaceTypeSpace, SpaceTypeClub:
		return true
	default:
		return false
	}
}
