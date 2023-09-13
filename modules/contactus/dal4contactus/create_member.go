package dal4contactus

import (
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/modules/memberus/briefs4memberus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dto4teamus"
	"github.com/strongo/validation"
)

var _ facade.Request = (*CreateMemberRequest)(nil)

// CreateMemberRequest request
type CreateMemberRequest struct {
	dto4teamus.TeamRequest
	briefs4memberus.MemberBase
	Relationship string `json:"relationship"` // Related to creator
	Message      string `json:"message"`
}

// Validate validates request
func (v *CreateMemberRequest) Validate() error {
	if err := v.TeamRequest.Validate(); err != nil {
		return err
	}
	if err := v.MemberBase.Validate(); err != nil {
		return err
	}
	// Validate relationship
	if v.Relationship != "" && !dbmodels.IsKnownRelationship(v.Relationship) {
		return validation.NewErrBadRequestFieldValue("relationship", "unknown value: "+v.Relationship)
	}
	return nil
}

// CreateTeamMemberResponse response
type CreateTeamMemberResponse struct {
	Member ContactContext `json:"member"`
}

// Validate returns error if not valid
func (v CreateTeamMemberResponse) Validate() error {
	if err := v.Member.Data.Validate(); err != nil {
		return validation.NewErrBadRecordFieldValue("member", err.Error())
	}
	return nil
}
