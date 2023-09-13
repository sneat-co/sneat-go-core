package dal4memberus

import (
	"errors"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core"
	"github.com/sneat-co/sneat-go-core/modules/contactus/models4contactus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dal4teamus"
)

// teamMembersCollection defines collection name
const teamMembersCollection = models4contactus.TeamContactsCollection //"members"

func NewFullMemberID(teamID, memberID string) string {
	return fmt.Sprintf("%v:%v", teamID, memberID)
}

// NewMemberKey creates a new member's key
func NewMemberKey(teamID, memberID string) *dal.Key {
	if err := ValidateMemberID(memberID); err != nil {
		panic(err)
	}
	//fullMemberID := NewFullMemberID(teamID, memberID)
	teamKey := dal4teamus.NewTeamKey(teamID)
	return dal.NewKeyWithParentAndID(teamKey, teamMembersCollection, memberID)
}

func ValidateMemberID(id string) error {
	if id == "" {
		return errors.New("empty member ContactID")
	}
	if !core.IsAlphanumericOrUnderscore(id) {
		return errors.New("member ContactID contains non alphanumeric characters")
	}
	return nil
}
