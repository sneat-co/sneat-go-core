package dbmodels

import (
	"fmt"
	"github.com/strongo/slice"
	"github.com/strongo/validation"
	"strings"
)

// WithMultiTeamAssetIDs mixin that adds indexed AssetIDs field // TODO: should be moved to assetus module?
type WithMultiTeamAssetIDs struct {
	// AssetIDs is used to indicate links to other assets for indexed search
	AssetIDs []string `json:"assetIDs,omitempty" firestore:"assetIDs,omitempty"`
}

// Validate  returns error if not valid
func (v WithMultiTeamAssetIDs) Validate() error {
	if len(v.AssetIDs) == 0 {
		return validation.NewErrRecordIsMissingRequiredField("assetIDs")
	}
	if v.AssetIDs[0] != "*" {
		return validation.NewErrBadRecordFieldValue("assetIDs[0]", "should be '*'")
	}
	for i, id := range v.AssetIDs[1:] {
		if strings.TrimSpace(id) == "" {
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("contactIDs[%v]", i), "can not be empty string")
		}
		ids := strings.Split(id, ":")
		if len(ids) != 2 {
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("contactIDs[%v]", i), "should be in format 'teamID:assetID', got: "+id)
		}
		if ids[0] == "" {
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("contactIDs[%v]", i), "teamID can not be empty string")
		}
		if ids[1] == "" {
			return validation.NewErrBadRecordFieldValue(fmt.Sprintf("contactIDs[%v]", i), "assetID can not be empty string")
		}
	}
	return nil
}

// HasAssetID check if a record has a specific contact ContactID
func (v WithMultiTeamAssetIDs) HasAssetID(id string) bool {
	if id == "*" {
		panic("id == '*'")
	}
	if id == "" {
		panic("id is empty string")
	}
	return slice.Contains(v.AssetIDs, id)
}
