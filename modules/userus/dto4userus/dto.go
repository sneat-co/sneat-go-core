package dto4userus

import (
	"fmt"
	"github.com/sneat-co/sneat-go-core"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dto4teamus"
	"github.com/sneat-co/sneat-go-core/validate"
	"github.com/strongo/validation"
	"net/mail"
	"strings"
)

var _ facade.Request = (*InitUserRecordRequest)(nil)

type InitTeamInfo struct {
	Type  string `json:"type"`
	Title string `json:"title"`
}

func (v InitTeamInfo) Validate() error {
	if strings.TrimSpace(v.Type) == "" {
		return validation.NewErrRequestIsMissingRequiredField("type")
	}
	if strings.TrimSpace(v.Title) == "" {
		return validation.NewErrRequestIsMissingRequiredField("title")
	}
	return nil
}

// InitUserRecordRequest request
type InitUserRecordRequest struct {
	AuthProvider    string                        `json:"authProvider,omitempty"` // TODO: Can we get it from Firebase token?
	Email           string                        `json:"email,omitempty"`
	EmailIsVerified bool                          `json:"emailIsVerified,omitempty"`
	IanaTimezone    string                        `json:"ianaTimezone,omitempty"`
	Name            *dbmodels.Name                `json:"name"`
	Team            *dto4teamus.CreateTeamRequest `json:"team,omitempty"`
	//
	RemoteClient dbmodels.RemoteClientInfo `json:"remoteClient"`
}

// Validate validates request
func (v *InitUserRecordRequest) Validate() error {
	if v.AuthProvider == "" {
		return validation.NewErrRequestIsMissingRequiredField("authProvider")
	} else if !dbmodels.IsKnownAuthProviderID(v.AuthProvider) {
		return validation.NewErrBadRequestFieldValue("authProvider", "unknown value: "+v.AuthProvider)
	}
	if v.Name != nil {
		if err := v.Name.Validate(); err != nil {
			return fmt.Errorf("%w: %v", facade.ErrBadRequest, err)
		}
	}
	if _, err := mail.ParseAddress(v.Email); err != nil {
		return validation.NewErrBadRequestFieldValue("email", err.Error())
	}
	if v.Team != nil {
		if err := v.Team.Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue("team", err.Error())
		}
	}
	return nil
}

// CreateUserRequest DTO
type CreateUserRequest struct {
	Creator string `json:"creator"`
	Title   string `json:"title,omitempty"`
	Email   string `json:"email"`
}

// Validate validates request
func (v *CreateUserRequest) Validate() error {
	//if err := validate.RequestTitle(v.Title, "title"); err != nil {
	//	return err
	//}
	if err := validate.OptionalEmail(v.Email, "email"); err != nil {
		return err
	}
	if v.Creator == "" {
		return validation.NewErrRecordIsMissingRequiredField("creator")
	} else if !core.IsKnownHost(v.Creator) {
		return validation.NewErrBadRequestFieldValue("creator", "unknown creator: "+v.Creator)
	}
	return nil
}

// CreateUserRequestWithRemoteClientInfo a request DTO
type CreateUserRequestWithRemoteClientInfo struct {
	CreateUserRequest
	RemoteClient dbmodels.RemoteClientInfo
}
