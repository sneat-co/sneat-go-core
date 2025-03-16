package coretypes

import (
	"fmt"
	"github.com/dal-go/dalgo/update"
	"github.com/strongo/validation"
	"slices"
	"strings"
)

type Permission string

type ShareDetail struct {
	ByUserID    string       `json:"byUserID" firestore:"byUserID"`
	Permissions []Permission `json:"permissions" firestore:"permissions"`
}

func (v ShareDetail) Validate() error {
	if v.ByUserID == "" {
		return validation.NewErrRequestIsMissingRequiredField("byUserID")
	}
	if len(v.Permissions) == 0 {
		return validation.NewErrRecordIsMissingRequiredField("permissions")
	}
	for i, p := range v.Permissions {
		if p == "" {
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("permissions[%d]", i), "is empty string")
		}
		if strings.TrimSpace(string(p)) != string(p) {
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("permissions[%d]", i),
				"has leading or trailing white space characters")
		}
	}
	return nil
}

// SharedToMap is map[EntityRefWithoutSpaceID]map[SpaceID]ShareDetail
type SharedToMap map[string]map[string]ShareDetail

// SharedFromMap is map[SpaceID]map[EntityRefWithoutSpaceID]ShareDetail
type SharedFromMap map[string]map[string]ShareDetail

type WithSharedMap struct {
	SharedTo   SharedToMap   `json:"sharedTo,omitempty" firestore:"sharedTo,omitempty"`
	SharedFrom SharedFromMap `json:"sharedFrom,omitempty" firestore:"sharedFrom,omitempty"`
}

func (v *WithSharedMap) Validate() error {
	if err := validateSharedTo("sharedTo", v.SharedTo); err != nil {
		return err
	}
	if err := validateSharedFrom("sharedFrom", v.SharedFrom); err != nil {
		return err
	}
	return nil
}

func (v *WithSharedMap) AddSharedTo(entityRef EntityRefWithoutSpaceID, spaceID string, shareDetail ShareDetail) (u update.Update, err error) {
	if entityRef == "" {
		panic("entityRef is a required parameter")
	}
	if spaceID == "" {
		panic("spaceID is a required parameter")
	}
	if v.SharedTo == nil {
		v.SharedTo = make(SharedToMap, 1)
	}
	spaces := v.SharedTo[string(entityRef)]
	if spaces == nil {
		spaces = make(map[string]ShareDetail, 1)
		v.SharedTo[string(entityRef)] = spaces
	}
	if existing, ok := spaces[spaceID]; ok && existing.ByUserID == shareDetail.ByUserID && slices.Equal(existing.Permissions, shareDetail.Permissions) {
		return nil, nil
	}
	spaces[spaceID] = shareDetail
	return update.ByFieldPath([]string{"sharedTo", string(entityRef), spaceID}, shareDetail), nil
}

func (v *WithSharedMap) AddSharedFrom(spaceID string, entityRef EntityRefWithoutSpaceID, shareDetail ShareDetail) (u update.Update, err error) {
	if spaceID == "" {
		panic("spaceID is a required parameter")
	}
	if entityRef == "" {
		panic("entityRef is a required parameter")
	}
	if v.SharedFrom == nil {
		v.SharedFrom = make(SharedFromMap, 1)
	}
	entities := v.SharedFrom[spaceID]
	if entities == nil {
		entities = make(map[string]ShareDetail, 1)
		v.SharedTo[string(entityRef)] = entities
	}
	if existing, ok := entities[string(entityRef)]; ok && existing.ByUserID == shareDetail.ByUserID && slices.Equal(existing.Permissions, shareDetail.Permissions) {
		return nil, nil
	}
	entities[spaceID] = shareDetail
	return update.ByFieldPath([]string{"sharedFrom", spaceID, string(entityRef)}, shareDetail), nil
}

func validateSharedFrom(fieldName string, sharedFrom SharedFromMap) (err error) {
	for spaceID, refs := range sharedFrom {
		if strings.TrimSpace(string(spaceID)) != string(spaceID) {
			return validation.NewErrBadRecordFieldValue("fieldName", "spaceID key has leading or trailing spaces")
		}
		for entityRef, shareDetail := range refs {
			if err = EntityRefWithoutSpaceID(entityRef).Validate(); err != nil {
				return validation.NewErrBadRecordFieldValue(fmt.Sprintf("%s.%s", fieldName, spaceID), err.Error())
			}
			if err = shareDetail.Validate(); err != nil {
				return validation.NewErrBadRecordFieldValue(fmt.Sprintf("%s.%s.%s", fieldName, spaceID, entityRef), err.Error())
			}
		}
	}
	return nil
}

func validateSharedTo(fieldName string, sharedTo SharedToMap) (err error) {
	for entityRef, refs := range sharedTo {
		if strings.TrimSpace(string(entityRef)) != string(entityRef) {
			return validation.NewErrBadRecordFieldValue("fieldName", "entityRef key has leading or trailing spaces")
		}
		if err = EntityRefWithoutSpaceID(entityRef).Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("%s.%s", fieldName, entityRef), err.Error())
		}
		for spaceID, shareDetail := range refs {
			if err = shareDetail.Validate(); err != nil {
				return validation.NewErrBadRecordFieldValue(fmt.Sprintf("%s.%s.%s", fieldName, entityRef, spaceID), err.Error())
			}
		}
	}
	return nil
}
