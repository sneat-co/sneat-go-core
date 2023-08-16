package dbmodels

import (
	"github.com/strongo/validation"
	"strings"
)

const (
	// RelationshipSpouse = "spouse"
	RelationshipSpouse = "spouse"

	// RelationshipPartner = "partner"
	RelationshipPartner = "partner"

	// RelationshipChild = "child"
	RelationshipChild = "child"

	// RelationshipSibling     = "sibling"
	RelationshipSibling = "sibling"

	// RelationshipParent      = "parent"
	RelationshipParent = "parent"

	// RelationshipGrandparent = "grandparent"
	RelationshipGrandparent = "grandparent"

	// RelationshipOther       = "other"
	RelationshipOther = "other"

	// RelationshipUnknown = "unknown"
	RelationshipUnknown = "unknown"

	// RelationshipUndisclosed = "undisclosed"
	RelationshipUndisclosed = "undisclosed"
)

// IsKnownRelationship checks if values is a known relationship
func IsKnownRelationship(v string) bool {
	switch v {
	case
		RelationshipOther,
		RelationshipUnknown,
		RelationshipUndisclosed,
		RelationshipParent,
		RelationshipChild,
		RelationshipSpouse,
		RelationshipPartner,
		RelationshipSibling,
		RelationshipGrandparent:
		return true
	}
	return false
}

const (
	// AgeGroupUnknown     = "unknown"
	AgeGroupUnknown = "unknown"

	// AgeGroupAdult       = "adult"
	AgeGroupAdult = "adult"

	// AgeGroupChild       = "child"
	AgeGroupChild = "child"

	// AgeGroupSenior      = "senior" // Should be removed?
	AgeGroupSenior = "senior" // Should be removed?

	// AgeGroupUndisclosed = "undisclosed"
	AgeGroupUndisclosed = "undisclosed"
)

// AgeGroups defines known age groups
var AgeGroups = []string{
	AgeGroupUnknown,
	AgeGroupAdult,
	AgeGroupChild,
	AgeGroupSenior,
	AgeGroupUndisclosed,
}

// ValidateAgeGroup return error if not a valid age group, TODO: similar but not same as IsKnownAgeGroupOrEmpty?
func ValidateAgeGroup(v string, required bool) error {
	if required && strings.TrimSpace(v) == "" {
		return validation.NewErrRecordIsMissingRequiredField("ageGroup")
	}
	if !IsKnownAgeGroupOrEmpty(v) {
		return validation.NewErrBadRecordFieldValue("ageGroup", "unknown value: "+v)
	}
	return nil
}

// IsKnownAgeGroupOrEmpty returns error if not a valid age group
func IsKnownAgeGroupOrEmpty(v string) bool {
	if v == "" {
		return true
	}
	for _, g := range AgeGroups {
		if g == v {
			return true
		}
	}
	return false
}
