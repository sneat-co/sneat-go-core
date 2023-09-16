package facade4memberus

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dal4contactus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/briefs4memberus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dto4teamus"
	"github.com/sneat-co/sneat-go-core/modules/userus/models4userus"
)

// LeaveTeam leaves team
func LeaveTeam(ctx context.Context, userContext facade.User, request dto4teamus.LeaveTeamRequest) (err error) {
	if err = request.Validate(); err != nil {
		return
	}
	return dal4contactus.RunContactusTeamWorker(ctx, userContext, request.TeamRequest,
		func(ctx context.Context, tx dal.ReadwriteTransaction, params *dal4contactus.ContactusTeamWorkerParams) (err error) {

			if err := tx.GetMulti(ctx, []dal.Record{params.ContactusTeam.Record}); err != nil {
				return fmt.Errorf("failed to get team record: %w", err)
			}

			uid := userContext.GetID()
			user := new(models4userus.UserDto)
			userKey := dal.NewKeyWithID(models4userus.UsersCollection, uid)
			userRecord := dal.NewRecordWithData(userKey, nil)
			if err = tx.Get(ctx, userRecord); err != nil {
				return
			}

			// Update team record
			{
				team := params.Team
				var updates []dal.Update
				var memberUserID string
				memberUserID, updates, err = removeTeamMember(team, params.ContactusTeam,
					func(_ string, m *briefs4memberus.MemberBrief) bool {
						return m.UserID == uid
					})
				if err != nil || len(updates) == 0 {
					return
				}
				if memberUserID != uid {
					err = fmt.Errorf("user ContactID does not match members record: memberUserID[%v] != ctx.UserID[%v]: %w",
						memberUserID, uid, facade.ErrBadRequest)
					return
				}
				if err = team.Data.Validate(); err != nil {
					return fmt.Errorf("team reacord is not valid: %v", err)
				}
				if len(params.ContactusTeam.Data.Contacts) == 0 || len(team.Data.UserIDs) == 0 {
					if err = tx.Delete(ctx, team.Key); err != nil {
						return err
					}
				} else {
					if err = txUpdateMemberGroup(ctx, tx, params.Started, params.Team.Data, params.Team.Key, updates); err != nil {
						return
					}
				}
			}

			// Update user
			{
				if userRecord.Exists() {
					update := updateUserRecordOnTeamMemberRemoved(user, request.TeamID)
					if update != nil {
						if err = txUpdate(ctx, tx, userKey, []dal.Update{*update}); err != nil {
							return err
						}
					}
				}
			}
			return err
		})
}
