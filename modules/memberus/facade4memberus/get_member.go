package facade4memberus

import (
	"context"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/modules/contactus/models4contactus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/dal4memberus"
)

// GetMemberByID returns member by ID
// Deprecated: use dal4contactus.GetContactByID() instead
func GetMemberByID(ctx context.Context, getter dal.ReadSession, teamID, memberID string) (memberDto *models4contactus.ContactDto, memberRecord dal.Record, err error) {
	memberKey := dal4memberus.NewMemberKey(teamID, memberID)
	memberDto = new(models4contactus.ContactDto)
	memberRecord = dal.NewRecordWithData(memberKey, memberDto)
	err = getter.Get(ctx, memberRecord)
	return memberDto, memberRecord, err
}
