package dto4contactus

import (
	"fmt"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/modules/contactus/briefs4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/models4contactus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dto4teamus"
	"github.com/strongo/validation"
)

// CreateContactRequest DTO
type CreateContactRequest struct {
	dto4teamus.TeamRequest
	dbmodels.WithRoles
	ParentContactID string                       `json:"parentContactID,omitempty"`
	Type            briefs4contactus.ContactType `json:"type"`
	Status          string                       `json:"status"`
	Person          *CreatePersonRequest         `json:"person,omitempty"`
	Company         *CreateCompanyRequest        `json:"company,omitempty"`
	Location        *CreateLocationRequest       `json:"location,omitempty"`
	Basic           *CreateBasicContactRequest   `json:"basic,omitempty"`

	// Used for situation when we want a hard-coded contact number
	// (e.g. a self-contact for a company team).
	// Can not be used from client side
	ContactID string `json:"-"`
}

func (v CreateContactRequest) Validate() error {
	if err := v.TeamRequest.Validate(); err != nil {
		return err
	}
	switch v.Status {
	case "":
		return validation.NewErrRequestIsMissingRequiredField("status")
	case "active", "draft":
		// OK
	default:
		return validation.NewErrBadRequestFieldValue("status", "allowed values are 'active' and 'draft', got: "+v.Status)
	}
	switch v.Type {
	case "":
		return validation.NewErrRequestIsMissingRequiredField("type")
	case "person":
		if v.Person == nil {
			return validation.NewErrRequestIsMissingRequiredField("person")
		}
		if err := v.Person.Validate(); err != nil {
			return validation.NewErrBadRequestFieldValue("person", fmt.Sprintf("contact type is set to 'c', but the `company` person is invalid: %v", err))
		}
	case "company":
		if v.Company == nil {
			return validation.NewErrRequestIsMissingRequiredField("company")
		}
		if err := v.Company.Validate(); err != nil {
			return validation.NewErrBadRequestFieldValue("company", fmt.Sprintf("contact type is set to 'company', but the `company` field is invalid: %v", err))
		}
	case "location":
		if v.Location == nil {
			return validation.NewErrRequestIsMissingRequiredField("location")
		}
		if err := v.Location.Validate(); err != nil {
			return validation.NewErrBadRequestFieldValue("location", fmt.Sprintf("contact type is set to 'location', but the `location` field is invalid: %v", err))
		}
		if v.ParentContactID == "" {
			return validation.NewErrRequestIsMissingRequiredField("parentContactID")
		}
	default:
		if v.Basic == nil {
			return validation.NewErrRequestIsMissingRequiredField("basic")
		}
		if err := v.Basic.Validate(); err != nil {
			return validation.NewErrBadRequestFieldValue("company", err.Error())
		}
	}
	if err := v.WithRoles.Validate(); err != nil {
		return fmt.Errorf("%w: %v", facade.ErrBadRequest, err.Error())
	}
	return nil
}

// CreateContactResponse DTO
type CreateContactResponse struct {
	ID  string                       `json:"id"`
	Dto *models4contactus.ContactDto `json:"dto,omitempty"`
}
