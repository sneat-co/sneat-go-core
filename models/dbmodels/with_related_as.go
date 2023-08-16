package dbmodels

import (
	"github.com/strongo/validation"
	"strings"
)

type RelatedAs interface {
	Validate() error
	GetRelatedAs() string
}

// WithOptionalRelatedAs is a mixin that adds RelatedAs field
type WithOptionalRelatedAs struct {
	RelatedAs string `json:"relatedAs,omitempty" firestore:"relatedAs,omitempty"`
}

// GetRelatedAs returns RelatedAs value - needed so we can pass as an interface
func (v WithOptionalRelatedAs) GetRelatedAs() string {
	return v.RelatedAs
}

func (v WithOptionalRelatedAs) Equal(v2 WithOptionalRelatedAs) bool {
	return v.RelatedAs == v2.RelatedAs
}

// Validate returns error if relatedAs has leading or trailing spaces
func (v WithOptionalRelatedAs) Validate() error {
	switch strings.TrimSpace(v.RelatedAs) {
	case "", v.RelatedAs: // OK
	default:
		return validation.NewErrBadRecordFieldValue("relatedAs", "has leading or trailing spaces")
	}
	return nil
}

// WithRequiredRelatedAs is a mixin that adds RelatedAs field
type WithRequiredRelatedAs struct {
	WithOptionalRelatedAs
}

// Validate returns error if relatedAs is empty
func (v WithRequiredRelatedAs) Validate() error {
	if err := v.WithOptionalRelatedAs.Validate(); err != nil {
		return err
	}
	if v.RelatedAs == "" {
		return validation.NewErrRecordIsMissingRequiredField("relatedAs")
	}
	return nil
}

func (v WithRequiredRelatedAs) Equal(v2 WithRequiredRelatedAs) bool {
	return v.RelatedAs == v2.RelatedAs
}
