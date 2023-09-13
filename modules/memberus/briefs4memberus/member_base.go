package briefs4memberus

import (
	"fmt"
	"github.com/sneat-co/sneat-go-core/modules/contactus/briefs4contactus"
	"github.com/strongo/validation"
)

// MemberBase DTO
type MemberBase struct {
	// MemberType string `json:"memberType"` - decided against it in favor of `roles`
	briefs4contactus.ContactBase
}

// Validate return error if not valid
func (v MemberBase) Validate() error {
	if err := v.ContactBase.Validate(); err != nil {
		return err
	}
	if len(v.Roles) == 0 {
		return validation.NewErrRecordIsMissingRequiredField("roles")
	}
	for i, role := range v.Roles {
		switch role {
		case "":
			field := fmt.Sprintf("roles[%v]", i)
			return validation.NewErrBadRecordFieldValue(field, "empty role")
		case
			TeamMemberRoleCreator,
			TeamMemberRoleOwner,
			TeamMemberRoleContributor,
			TeamMemberRoleSpectator,
			TeamMemberRoleScrumMaster:
			break
		default:
			return validation.NewErrBadRecordFieldValue("Role",
				fmt.Sprintf("unknown member role='%v' (expected one of: %v, %v %v)", role,
					TeamMemberRoleContributor, TeamMemberRoleSpectator, TeamMemberRoleScrumMaster))
		}
	}
	return nil
}
