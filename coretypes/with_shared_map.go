package coretypes

import (
	"fmt"
	"github.com/strongo/validation"
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

// SharedMap is map[SpaceID]map[EntityRefWithoutSpaceID]ShareDetail
type SharedMap map[string]map[string]ShareDetail

type WithSharedMap struct {
	SharedTo   SharedMap `json:"sharedTo,omitempty" firestore:"sharedTo,omitempty"`
	SharedFrom SharedMap `json:"sharedFrom,omitempty" firestore:"sharedFrom,omitempty"`
}

func (v WithSharedMap) Validate() error {
	validate := func(fieldName string, shared SharedMap) error {
		for spaceID, refs := range shared {
			if strings.TrimSpace(string(spaceID)) != string(spaceID) {
				return validation.NewErrBadRecordFieldValue("fieldName", "spaceID key has leading or trailing spaces")
			}
			for entityRef, shareDetail := range refs {
				if err := EntityRefWithoutSpaceID(entityRef).Validate(); err != nil {
					return validation.NewErrBadRecordFieldValue(fmt.Sprintf("%s.%s", fieldName, spaceID), err.Error())
				}
				if err := shareDetail.Validate(); err != nil {
					return validation.NewErrBadRecordFieldValue(fmt.Sprintf("%s.%s.%s", fieldName, spaceID, entityRef), err.Error())
				}
			}
		}
		return nil
	}
	if err := validate("sharedTo", v.SharedTo); err != nil {
		return err
	}
	if err := validate("sharedFrom", v.SharedFrom); err != nil {
		return err
	}
	// we can have reciprocal shares for calendar events when an event can be edit by a member of either space
	// Consider if we should prevent this happening in other cases?
	//for spaceID := range v.SharedTo {
	//	if _, ok := v.SharedFrom[spaceID]; ok {
	//		return validation.NewErrBadRecordFieldValue("sharedFrom&sharedTo", "same space ID in sharedTo and sharedFrom: "+spaceID)
	//	}
	//}
	return nil
}
