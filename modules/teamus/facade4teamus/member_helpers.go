package facade4teamus

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dal4contactus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/briefs4memberus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/dal4memberus"
)

// CreateMemberRecordFromBrief creates a member record from member's brief
func CreateMemberRecordFromBrief(
	ctx context.Context,
	tx dal.ReadwriteTransaction,
	teamID string,
	contactID string,
	memberBrief briefs4memberus.MemberBrief,
) (
	member dal4memberus.MemberContext,
	err error,
) {
	if err = memberBrief.Validate(); err != nil {
		return member, fmt.Errorf("supplied member brief is not valid: %w", err)
	}
	member = dal4contactus.NewContactContext(teamID, contactID)
	//member.Brief = &memberBrief
	//member.Data.TeamID = teamID
	member.Data.ContactBrief = memberBrief
	_ = member.Data.AddRole(briefs4memberus.TeamMemberRoleTeamMember)
	if err := tx.Insert(ctx, member.Record); err != nil {
		return member, fmt.Errorf("failed to inser member record into DB: %w", err)
	}
	return member, nil
}
