package facade4memberus

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dal4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dto4contactus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/briefs4memberus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dal4teamus"
	"github.com/sneat-co/sneat-go-core/modules/userus/facade4userus"
	"github.com/sneat-co/sneat-go-core/modules/userus/models4userus"
	"github.com/strongo/slice"
)

// RemoveMember removes members from a team
func RemoveMember(ctx context.Context, user facade.User, request dto4contactus.ContactRequest) (err error) {
	if err = request.Validate(); err != nil {
		return err
	}
	return dal4contactus.RunContactusTeamWorker(ctx, user, request.TeamRequest,
		func(ctx context.Context, tx dal.ReadwriteTransaction, params *dal4contactus.ContactusTeamWorkerParams) (err error) {
			var updates []dal.Update
			var memberUserID string

			if memberUserID, updates, err = removeTeamMember(params.Team, params.ContactusTeam,
				func(contactID string, _ *briefs4memberus.MemberBrief) bool {
					return contactID == request.ContactID
				},
			); err != nil || len(updates) == 0 {
				return
			}

			if memberUserID != "" {
				var (
					userRef *dal.Key
				)
				user := new(models4userus.UserDto)
				userRecord := dal.NewRecordWithData(models4userus.NewUserKey(memberUserID), user)
				if err = facade4userus.TxGetUserByID(ctx, tx, userRecord); err != nil {
					return
				}

				update := updateUserRecordOnTeamMemberRemoved(user, request.TeamID)
				if update != nil {
					if err = txUpdate(ctx, tx, userRef, []dal.Update{*update}); err != nil {
						return err
					}
				}
			}
			if err = params.Team.Data.Validate(); err != nil {
				return fmt.Errorf("team reacord is not valid: %v", err)
			}
			if err = txUpdateMemberGroup(ctx, tx, params.Started, params.Team.Data, params.Team.Key, updates); err != nil {
				return
			}
			return
		})
}

func updateUserRecordOnTeamMemberRemoved(user *models4userus.UserDto, teamID string) *dal.Update {
	delete(user.Teams, teamID)
	user.TeamIDs = slice.RemoveInPlace(teamID, user.TeamIDs)
	return &dal.Update{
		Field: "teams",
		Value: user.Teams,
	}
}

func removeTeamMember(
	team dal4teamus.TeamContext,
	contactusTeam dal4contactus.ContactusTeamContext,
	match func(contactID string, m *briefs4memberus.MemberBrief) bool,
) (memberUserID string, updates []dal.Update, err error) {
	userIds := contactusTeam.Data.UserIDs

	for id, m := range contactusTeam.Data.Contacts {
		if match(id, m) {
			if m.UserID != "" {
				memberUserID = m.UserID
				userIds = removeTeamUserID(userIds, m.UserID)
			}
			updates = append(updates, contactusTeam.Data.RemoveContact(id))
		}
	}
	if len(userIds) != len(contactusTeam.Data.UserIDs) {
		contactusTeam.Data.UserIDs = userIds
		if len(userIds) == 0 {
			userIds = nil
		}
		updates = []dal.Update{
			{Field: "userIDs", Value: userIds},
		}
	}
	updates = append(updates, team.Data.SetNumberOf("contacts", len(contactusTeam.Data.Contacts)))
	updates = append(updates, team.Data.SetNumberOf("members", len(contactusTeam.Data.GetContactBriefsByRoles(briefs4memberus.TeamMemberRoleTeamMember))))
	return
}

func removeTeamUserID(userIDs []string, userID string) []string {
	uIDs := make([]string, 0, len(userIDs))
	for _, uid := range userIDs {
		if uid != userID {
			uIDs = append(uIDs, uid)
		}
	}
	return uIDs
}
