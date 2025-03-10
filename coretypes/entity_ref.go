package coretypes

import (
	"github.com/strongo/validation"
	"strings"
)

const EntityRefSeparator = ":"
const EntitySpaceSeparator = "@"

type EntityRef string
type EntityRefWithSpaceID = EntityRef
type EntityRefWithoutSpaceID = EntityRef

func NewEntityRef(kind, entityID string) EntityRef {
	if kind == "" {
		panic("kind is a required parameter")
	}
	if entityID == "" {
		panic("entityID is a required parameter")
	}
	return EntityRef(kind + EntityRefSeparator + entityID)
}

func NewEntityRefWithSpaceID(kind, entityID, spaceID string) EntityRef {
	if spaceID == "" {
		panic("spaceID is a required argument")
	}
	return NewEntityRef(kind, entityID) + EntityRef(EntitySpaceSeparator+spaceID)
}

func (v EntityRef) Kind() string {
	i := strings.Index(string(v), EntityRefSeparator)
	if i >= 0 {
		return string(v)[:i]
	}
	return ""
}

func (v EntityRef) ID() string {
	idIdx := strings.Index(string(v), EntityRefSeparator)
	if idIdx < 0 {
		return ""
	}
	if spaceIdx := strings.Index(string(v), EntitySpaceSeparator); spaceIdx >= 0 {
		return string(v)[idIdx+1 : spaceIdx]
	}
	return string(v)[idIdx+1:]
}
func (v EntityRef) AddSpaceID(spaceID SpaceID) EntityRef {
	if spaceID == "" {
		panic("spaceID is a required parameter")
	}
	return EntityRef(string(v) + EntitySpaceSeparator + string(spaceID))
}

func (v EntityRef) SpaceID() SpaceID {
	i := strings.Index(string(v), EntitySpaceSeparator)
	if i < 0 {
		return ""
	}
	return SpaceID(string(v)[i+1:])
}

func (v EntityRef) Validate() error {
	if v == "" {
		return validation.NewErrRequestIsMissingRequiredField("entityRef")
	}
	if strings.TrimSpace(string(v)) == "" {
		return validation.NewErrBadRecordFieldValue("entityRef", "empty value")
	}
	if strings.TrimSpace(string(v)) != string(v) {
		return validation.NewErrBadRecordFieldValue("entityRef", "has leading or trailing spaces")
	}
	switch i := strings.Index(string(v), EntityRefSeparator); i {
	case 0:
		return validation.NewErrBadRecordFieldValue("entityRef", "missing kind: "+string(v))
	case -1:
		return validation.NewErrBadRecordFieldValue("entityRef", "missing separator: "+string(v))
	case len(v) - 1:
		return validation.NewErrBadRecordFieldValue("entityRef", "missing userID: "+string(v))
	}
	if v.ID() == "" {
		return validation.NewErrBadRecordFieldValue("entityRef", "missing userID: "+string(v))
	}
	return nil
}
