package facade4invitus

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/modules/contactus/briefs4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dal4contactus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/briefs4memberus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/dal4memberus"
	"github.com/sneat-co/sneat-go-core/modules/userus/facade4userus"
	"github.com/sneat-co/sneat-go-core/modules/userus/models4userus"
	"github.com/strongo/validation"
	"strings"
	"time"
)

// AcceptPersonalInviteRequest holds parameters for accepting a personal invite
type AcceptPersonalInviteRequest struct {
	InviteRequest
	RemoteClient dbmodels.RemoteClientInfo `json:"remoteClient"`
	MemberID     string
	Member       dbmodels.DtoWithID[*briefs4contactus.ContactBase] `json:"member"`
	//FullName string                      `json:"fullName"`
	//Email    string                      `json:"email"`
}

// Validate validates request
func (v *AcceptPersonalInviteRequest) Validate() error {
	if err := v.InviteRequest.Validate(); err != nil {
		return err
	}
	if err := v.Member.Validate(); err != nil {
		return validation.NewErrBadRequestFieldValue("member", err.Error())
	}
	//if v.FullName == "" {
	//	return validation.NewErrRecordIsMissingRequiredField("FullName")
	//}
	//if v.Email == "" {
	//	return validation.NewErrRecordIsMissingRequiredField("Email")
	//}
	return nil
}

// AcceptPersonalInvite accepts personal invite and joins user to a team.
// If needed a user record should be created
func AcceptPersonalInvite(ctx context.Context, userContext facade.User, request AcceptPersonalInviteRequest) (err error) {
	if err = request.Validate(); err != nil {
		return err
	}
	uid := userContext.GetID()

	return dal4contactus.RunContactusTeamWorker(ctx, userContext, request.TeamRequest,
		func(ctx context.Context, tx dal.ReadwriteTransaction, params *dal4contactus.ContactusTeamWorkerParams) error {
			invite, member, err := getPersonalInviteRecords(ctx, tx, params, request.InviteID, request.Member.ID)
			if err != nil {
				return err
			}
			if invite.Dto.Status != "active" {
				return fmt.Errorf("invite status is not equal to 'active', got '%v'", invite.Dto.Status)
			}

			if invite.Dto.Pin != request.Pin {
				return fmt.Errorf("%w: pin code does not match", facade.ErrBadRequest)
			}

			user := models4userus.NewUserContext(uid)
			if err = facade4userus.GetUserByID(ctx, tx, user.Record); err != nil {
				if !dal.IsNotFound(err) {
					return err
				}
			}

			now := time.Now()

			if err = updateInviteRecord(ctx, tx, uid, now, invite, "accepted"); err != nil {
				return fmt.Errorf("failed to update invite record: %w", err)
			}

			var teamMember *briefs4contactus.ContactBase
			if teamMember, err = updateTeamRecord(uid, invite.Dto.ToTeamMemberID, params, request.Member); err != nil {
				return fmt.Errorf("failed to update team record: %w", err)
			}

			memberContext := dal4contactus.NewContactContext(params.Team.ID, member.ID)

			if err = updateMemberRecord(ctx, tx, uid, memberContext, request.Member.Data, teamMember); err != nil {
				return fmt.Errorf("failed to update team member record: %w", err)
			}

			if err = createOrUpdateUserRecord(ctx, tx, now, user, request, params, teamMember, invite); err != nil {
				return fmt.Errorf("failed to create or update user record: %w", err)
			}

			return err
		})
}

func updateInviteRecord(
	ctx context.Context,
	tx dal.ReadwriteTransaction,
	uid string,
	now time.Time,
	invite PersonalInviteContext,
	status string,
) (err error) {
	invite.Dto.Status = status
	invite.Dto.To.UserID = uid
	inviteUpdates := []dal.Update{
		{Field: "status", Value: status},
		{Field: "to.userID", Value: uid},
	}
	switch status {
	case "active":
		if invite.Dto.Claimed != nil {
			invite.Dto.Claimed = nil
			inviteUpdates = append(inviteUpdates, dal.Update{Field: "claimed", Value: dal.DeleteField})
		}
	case "expired": // Do nothing
	default:
		invite.Dto.Claimed = &now
		inviteUpdates = append(inviteUpdates, dal.Update{Field: "claimed", Value: now})
	}
	if err := invite.Dto.Validate(); err != nil {
		return fmt.Errorf("personal invite record is not valid: %w", err)
	}
	if err = tx.Update(ctx, invite.Key, inviteUpdates); err != nil {
		return err
	}
	return err
}

func updateMemberRecord(
	ctx context.Context,
	tx dal.ReadwriteTransaction,
	uid string,
	member dal4memberus.MemberContext,
	requestMember *briefs4contactus.ContactBase,
	teamMember *briefs4contactus.ContactBase,
) (err error) {
	updates := []dal.Update{
		{Field: "userID", Value: uid},
	}
	updates = updatePersonDetails(&member.Data.ContactBase, requestMember, teamMember, updates)
	if err = tx.Update(ctx, member.Key, updates); err != nil {
		return err
	}
	return err
}

func updateTeamRecord(
	uid, memberID string,
	params *dal4contactus.ContactusTeamWorkerParams,
	requestMember dbmodels.DtoWithID[*briefs4contactus.ContactBase],
) (teamMember *briefs4contactus.ContactBase, err error) {
	if uid == "" {
		panic("required parameter `uid` is empty string")
	}

	inviteToMemberID := memberID[strings.Index(memberID, ":")+1:]
	for contactID, m := range params.ContactusTeam.Data.Contacts {
		if contactID == inviteToMemberID {
			m.UserID = uid
			params.ContactusTeam.Data.AddUserID(uid)
			params.ContactusTeam.Data.AddContact(contactID, m)
			//request.ContactID.Roles = m.Roles
			//m = request.ContactID
			m.UserID = uid
			teamMember = &briefs4contactus.ContactBase{
				ContactBrief: *m,
			}
			//team.Members[i] = m
			updatePersonDetails(teamMember, requestMember.Data, teamMember, nil)
			if u, ok := params.ContactusTeam.Data.AddUserID(uid); ok {
				params.ContactusTeamUpdates = append(params.ContactusTeamUpdates, u)
			}
			if m.AddRole(briefs4memberus.TeamMemberRoleTeamMember) {
				params.ContactusTeamUpdates = append(params.ContactusTeamUpdates, dal.Update{Field: "contacts." + contactID + ".roles", Value: m.Roles})
			}
			break
		}
	}
	if teamMember == nil {
		return teamMember, fmt.Errorf("team member is not found by ContactID=%v", inviteToMemberID)
	}

	if params.Team.Data.HasUserID(uid) {
		goto UserIdAdded
	}
	params.TeamUpdates = append(params.TeamUpdates, dal.Update{Field: "userIDs", Value: params.Team.Data.UserIDs})
UserIdAdded:
	return teamMember, err
}

func createOrUpdateUserRecord(
	ctx context.Context,
	tx dal.ReadwriteTransaction,
	now time.Time,
	user models4userus.UserContext,
	request AcceptPersonalInviteRequest,
	params *dal4contactus.ContactusTeamWorkerParams,
	teamMember *briefs4contactus.ContactBase,
	invite PersonalInviteContext,
) (err error) {
	if teamMember == nil {
		panic("teamMember == nil")
	}
	existingUser := user.Record.Exists()
	if existingUser {
		teamInfo := user.Dto.GetUserTeamInfoByID(request.TeamID)
		if teamInfo != nil {
			return nil
		}
	}

	userTeamInfo := models4userus.UserTeamBrief{
		TeamBrief: params.Team.Data.TeamBrief,
		Roles:     invite.Dto.Roles, // TODO: Validate roles?
	}
	if err = userTeamInfo.Validate(); err != nil {
		return fmt.Errorf("invalid user team info: %w", err)
	}
	user.Dto.Teams[request.TeamID] = &userTeamInfo
	user.Dto.TeamIDs = append(user.Dto.TeamIDs, request.TeamID)
	if existingUser {
		userUpdates := []dal.Update{
			{
				Field: "teams",
				Value: user.Dto.Teams,
			},
		}
		userUpdates = updatePersonDetails(&user.Dto.ContactBase, request.Member.Data, teamMember, userUpdates)
		if err = user.Dto.Validate(); err != nil {
			return fmt.Errorf("user record prepared for update is not valid: %w", err)
		}
		if err = tx.Update(ctx, user.Key, userUpdates); err != nil {
			return fmt.Errorf("failed to update user record: %w", err)
		}
	} else { // New user record
		user.Dto.Created.At = now
		user.Dto.Created.Client = request.RemoteClient
		user.Dto.Type = briefs4contactus.ContactTypePerson
		user.Dto.Name = request.Member.Data.Name
		if user.Dto.Name.IsEmpty() && teamMember != nil {
			user.Dto.Name = teamMember.Name
		}
		updatePersonDetails(&user.Dto.ContactBase, request.Member.Data, teamMember, nil)
		if user.Dto.Gender == "" {
			user.Dto.Gender = "unknown"
		}
		if user.Dto.CountryID == "" {
			user.Dto.CountryID = "--"
		}
		if len(request.Member.Data.Emails) > 0 {
			user.Dto.Emails = request.Member.Data.Emails
		}
		if len(request.Member.Data.Phones) > 0 {
			user.Dto.Phones = request.Member.Data.Phones
		}
		if err = user.Dto.Validate(); err != nil {
			return fmt.Errorf("user record prepared for insert is not valid: %w", err)
		}
		if err = tx.Insert(ctx, user.Record); err != nil {
			return fmt.Errorf("failed to insert user record: %w", err)
		}
	}
	return err
}

func updatePersonDetails(person *briefs4contactus.ContactBase, member *briefs4contactus.ContactBase, teamMember *briefs4contactus.ContactBase, updates []dal.Update) []dal.Update {
	if member.Name != nil {
		if person.Name == nil {
			person.Name = &dbmodels.Name{}
		}
		if person.Name.First == "" {
			name := member.Name.First
			if name == "" {
				name = teamMember.Name.First
			}
			if name != "" {
				person.Name.First = name
				if updates != nil {
					updates = append(updates, dal.Update{
						Field: "name.first",
						Value: name,
					})
				}
			}
		}
		if person.Name.Last == "" {
			name := member.Name.Last
			if name == "" {
				name = teamMember.Name.Last
			}
			if name != "" {
				person.Name.Last = name
				if updates != nil {
					updates = append(updates, dal.Update{
						Field: "name.last",
						Value: name,
					})
				}
			}
		}
		if person.Name.Full == "" {
			name := member.Name.Full
			if name == "" {
				name = teamMember.Name.Full
			}
			if name != "" {
				person.Name.Full = name
				if updates != nil {
					updates = append(updates, dal.Update{
						Field: "name.full",
						Value: name,
					})
				}
			}
		}
	}
	if person.Gender == "" || person.Gender == "unknown" {
		gender := member.Gender
		if gender == "" || gender == "unknown" {
			gender = teamMember.Gender
		}
		if gender == "" {
			gender = "unknown"
		}
		if person.Gender == "" || gender != "unknown" {
			person.Gender = member.Gender
			if updates != nil {
				updates = append(updates, dal.Update{
					Field: "gender",
					Value: gender,
				})
			}
		}
	}
	return updates
}
