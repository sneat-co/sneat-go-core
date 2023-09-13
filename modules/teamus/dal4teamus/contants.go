package dal4teamus

import (
	"github.com/dal-go/dalgo/dal"
)

const Collection = "modules"

func NewTeamModuleKey(teamID, moduleID string) *dal.Key {
	teamKey := NewTeamKey(teamID)
	return dal.NewKeyWithParentAndID(teamKey, Collection, moduleID)
}
