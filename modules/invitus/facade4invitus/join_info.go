package facade4invitus

import (
	"context"
	"errors"
	"fmt"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/modules/contactus/models4contactus"
	"github.com/sneat-co/sneat-go-core/modules/invitus/models4invitus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/briefs4memberus"
	"github.com/sneat-co/sneat-go-core/modules/memberus/facade4memberus"
	"github.com/strongo/validation"
	"strconv"
	"time"
)

// JoinInfoRequest request
type JoinInfoRequest struct {
	InviteID string `json:"inviteID"` // InviteDto ContactID
	Pin      string `json:"pin"`
}

// Validate validates request
func (v *JoinInfoRequest) Validate() error {
	if v.InviteID == "" {
		return validation.NewErrRecordIsMissingRequiredField("id")
	}
	if v.Pin == "" {
		return validation.NewErrRequestIsMissingRequiredField("pin")
	}
	if _, err := strconv.Atoi(v.Pin); err != nil {
		return validation.NewErrBadRequestFieldValue("pin", "%pin is expected to be an integer")
	}
	return nil
}

type InviteInfo struct {
	Created time.Time                 `json:"created"`
	Status  string                    `json:"status"`
	From    models4invitus.InviteFrom `json:"from"`
	To      *models4invitus.InviteTo  `json:"to"`
	Message string                    `json:"message,omitempty"`
}

func (v InviteInfo) Validate() error {
	if v.Status == "" {
		return validation.NewErrRecordIsMissingRequiredField("status")
	}
	if v.Created.IsZero() {
		return validation.NewErrRecordIsMissingRequiredField("created")
	}
	if err := v.From.Validate(); err != nil {
		return validation.NewErrBadRecordFieldValue("from", err.Error())
	}
	if err := v.To.Validate(); err != nil {
		return validation.NewErrBadRecordFieldValue("to", err.Error())
	}
	return nil
}

// JoinInfoResponse response
type JoinInfoResponse struct {
	Team   models4invitus.InviteTeam                         `json:"team"`
	Invite InviteInfo                                        `json:"invite"`
	Member *dbmodels.DtoWithID[*briefs4memberus.MemberBrief] `json:"member"`
}

func (v JoinInfoResponse) Validated() error {
	if err := v.Team.Validate(); err != nil {
		return validation.NewErrBadRecordFieldValue("team", err.Error())
	}
	if err := v.Invite.Validate(); err != nil {
		return validation.NewErrBadRecordFieldValue("team", err.Error())
	}
	if nil == v.Member {
		return validation.NewErrRecordIsMissingRequiredField("member")
	}
	if err := v.Member.Validate(); err != nil {
		return validation.NewErrBadRecordFieldValue("member", err.Error())
	}
	return nil
}

// GetTeamJoinInfo return join info
func GetTeamJoinInfo(ctx context.Context, request JoinInfoRequest) (response JoinInfoResponse, err error) {
	if err = request.Validate(); err != nil {
		return
	}
	db := facade.GetDatabase(ctx)

	var inviteDto *models4invitus.InviteDto
	inviteDto, _, err = GetInviteByID(ctx, db, request.InviteID)
	if err != nil {
		err = fmt.Errorf("failed to get invite record by ContactID=%v: %w", request.InviteID, err)
		return
	}
	if inviteDto == nil {
		err = errors.New("invite record not found by ContactID: " + request.InviteID)
		return
	}

	if inviteDto.Pin != request.Pin {
		err = fmt.Errorf("%v: %w",
			validation.NewErrBadRequestFieldValue("pin", "invalid pin"),
			facade.ErrForbidden,
		)
		return
	}
	var memberDto *models4contactus.ContactDto
	if inviteDto.To.MemberID != "" {
		memberDto, _, err = facade4memberus.GetMemberByID(ctx, db, inviteDto.TeamID, inviteDto.To.MemberID)
		if err != nil {
			err = fmt.Errorf("failed to get memer by ContactID: %w", err)
			return
		}
	}
	response.Team = inviteDto.Team
	response.Team.ID = inviteDto.TeamID
	response.Invite.Status = inviteDto.Status
	response.Invite.Created = inviteDto.Created.At
	response.Invite.From = inviteDto.From
	response.Invite.To = inviteDto.To
	response.Invite.Message = inviteDto.Message
	if inviteDto.To.MemberID != "" {
		response.Member = &dbmodels.DtoWithID[*briefs4memberus.MemberBrief]{
			ID:   inviteDto.To.MemberID,
			Data: &memberDto.ContactBrief,
		}
	}
	return response, nil
}