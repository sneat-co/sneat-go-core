package dbmodels

import (
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/validation"
	"strings"
)

// WithUserIDs defines a record with a list of user IDs
type WithUserIDs struct {
	UserIDs []string `json:"userIDs,omitempty" firestore:"userIDs,omitempty"`
}

// Validate returns error as soon as one of the fields is not valid
func (v *WithUserIDs) Validate() error {
	if len(v.UserIDs) == 0 {
		return validation.NewErrRecordIsMissingRequiredField("userIDs")
	}
	for i, uid := range v.UserIDs {
		if strings.TrimSpace(uid) == "" {
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("userIDs[%v]", i), "can not be empty string")
		}
	}
	return nil
}

// HasUserID checks if record has UserID
func (v *WithUserIDs) HasUserID(uid string) bool {
	for _, id := range v.UserIDs {
		if id == uid {
			return true
		}
	}
	return false
}

// AddUserID adds user ID and return dal.Update
func (v *WithUserIDs) AddUserID(uid string) (updates []dal.Update) {
	if v.HasUserID(uid) {
		return nil
	}
	v.UserIDs = append(v.UserIDs, uid)
	return []dal.Update{{
		Field: "userIDs",
		Value: v.UserIDs,
	}}
}
