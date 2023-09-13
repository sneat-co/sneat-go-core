package dal4contactus

import (
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core"
	"github.com/sneat-co/sneat-go-core/modules/contactus/models4contactus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dal4teamus"
)

// NewContactKey creates a new contact's key in format "teamID:memberID"
func NewContactKey(teamID, contactID string) *dal.Key {
	if !core.IsAlphanumericOrUnderscore(contactID) {
		panic(fmt.Errorf("contactID should be alphanumeric, got: [%v]", contactID))
	}
	teamKey := dal4teamus.NewTeamKey(teamID)
	return dal.NewKeyWithParentAndID(teamKey, models4contactus.TeamContactsCollection, contactID)
}
