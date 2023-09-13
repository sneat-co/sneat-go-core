package facade4invitus

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dal4contactus"
	"github.com/sneat-co/sneat-go-core/modules/invitus/models4invitus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/briefs4memberus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/dal4memberus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dto4teamus"
	"github.com/strongo/validation"
)

// GetPersonalInviteRequest holds parameters for creating a personal invite
type GetPersonalInviteRequest struct {
	dto4teamus.TeamRequest
	InviteID string `json:"inviteID"`
}

// Validate validates request
func (v *GetPersonalInviteRequest) Validate() error {
	if err := v.TeamRequest.Validate(); err != nil {
		return err
	}
	if v.InviteID == "" {
		return validation.NewErrRecordIsMissingRequiredField("invite")
	}
	//if len(v.InviteID) != 8 {
	//	return models2spotbuddies.NewErrBadRequestFieldValue("invite", "unexpected length of invite id")
	//}
	return nil
}

// PersonalInviteResponse holds response data for created personal invite
type PersonalInviteResponse struct {
	Invite  *models4invitus.PersonalInviteDto       `json:"invite,omitempty"`
	Members map[string]*briefs4memberus.MemberBrief `json:"members,omitempty"`
}

func getPersonalInviteRecords(ctx context.Context, getter dal.ReadSession, params *dal4contactus.ContactusTeamWorkerParams, inviteID, memberID string) (
	invite PersonalInviteContext,
	member dal4memberus.MemberContext,
	err error,
) {
	if inviteID == "" {
		err = validation.NewErrRequestIsMissingRequiredField("inviteID")
		return
	}
	invite = NewPersonalInviteContext(inviteID)

	records := []dal.Record{params.ContactusTeam.Record, invite.Record}
	if memberID != "" {
		member = dal4memberus.NewMemberContext(params.Team.ID, memberID)
		records = append(records, member.Record)
	}
	if err = getter.GetMulti(ctx, records); err != nil {
		return
	}
	if !params.ContactusTeam.Record.Exists() {
		err = validation.NewErrBadRequestFieldValue("teamID",
			fmt.Sprintf("contactusTeam record not found by key=%v: record.Error=%v",
				params.ContactusTeam.Key, params.ContactusTeam.Record.Error()),
		)
		return
	}
	if !invite.Record.Exists() {
		err = validation.NewErrBadRequestFieldValue("inviteID",
			fmt.Sprintf("invite record not found in database by key=%v: record.Error=%v",
				invite.Record.Key(), invite.Record.Error()))
		return
	}
	if member.Record != nil && !member.Record.Exists() {
		err = validation.NewErrBadRequestFieldValue("memberID",
			fmt.Sprintf("member record not found in database by key=%v: record.Error=%v",
				member.Record.Key(), member.Record.Error()))
		return
	}
	return
}

// GetPersonal returns personal invite data
func GetPersonal(ctx context.Context, user facade.User, request GetPersonalInviteRequest) (response PersonalInviteResponse, err error) {
	if err = request.Validate(); err != nil {
		return
	}
	return response, dal4contactus.RunReadonlyContactusTeamWorker(ctx, user, request.TeamRequest, func(ctx context.Context, tx dal.ReadTransaction, params *dal4contactus.ContactusTeamWorkerParams) error {
		invite, _, err := getPersonalInviteRecords(ctx, tx, params, request.InviteID, "")
		if err != nil {
			return err
		}
		invite.Dto.Pin = "" // Hide PIN code from visitor
		response = PersonalInviteResponse{
			Invite:  invite.Dto,
			Members: make(map[string]*briefs4memberus.MemberBrief, len(params.ContactusTeam.Data.Contacts)),
		}
		// TODO: Is this is a security breach in current implementation?
		//for id, contact := range contactusTeam.Data.Contacts {
		//	response.Members[id] = contact
		//}
		return nil
	})
}
