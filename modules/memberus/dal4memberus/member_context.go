package dal4memberus

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dal4contactus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/models4memberus"
)

type MemberContext = dal4contactus.ContactContext

func NewMemberContext(teamID, memberID string) MemberContext {
	return NewMemberContextWithDto(teamID, memberID, new(models4memberus.MemberDto))
}

func NewMemberContextWithDto(teamID, memberID string, dto *models4memberus.MemberDto) (member MemberContext) {
	member.ID = memberID
	member.FullID = NewFullMemberID(teamID, memberID)
	member.Key = NewMemberKey(teamID, memberID)
	member.Data = dto
	member.Record = dal.NewRecordWithData(member.Key, dto)
	return
}
