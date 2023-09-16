package facade4contactus

import (
	"context"
	"errors"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/modules/contactus/briefs4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/const4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dal4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dto4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/models4contactus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dal4teamus"
)

// CreateContact creates team contact
func CreateContact(
	ctx context.Context,
	userContext facade.User,
	request dto4contactus.CreateContactRequest,
) (
	response dto4contactus.CreateContactResponse,
	err error,
) {
	if err = request.Validate(); err != nil {
		return
	}

	err = dal4teamus.CreateTeamItem(ctx, userContext, "contacts", request.TeamRequest, const4contactus.ModuleID,
		func(ctx context.Context, tx dal.ReadwriteTransaction, params *dal4teamus.ModuleTeamWorkerParams[*models4contactus.ContactusTeamDto]) (err error) {
			var contact dal4contactus.ContactContext
			if contact, err = CreateContactTx(ctx, tx, params.UserID, request, params); err != nil {
				return err
			}
			response = dto4contactus.CreateContactResponse{
				ID:  contact.ID,
				Dto: contact.Data,
			}
			return err
		},
	)
	if err != nil {
		err = fmt.Errorf("failed to create a new contact: %w", err)
		return
	}
	return
}

func CreateContactTx(
	ctx context.Context,
	tx dal.ReadwriteTransaction,
	userID string,
	request dto4contactus.CreateContactRequest,
	params *dal4teamus.ModuleTeamWorkerParams[*models4contactus.ContactusTeamDto],
) (
	contact dal4contactus.ContactContext,
	err error,
) {
	if userID == "" {
		return contact, errors.New("only authenticated users can create team contacts")
	}
	if err = request.Validate(); err != nil {
		return
	}

	parentContactID := request.ParentContactID

	var parent dal4contactus.ContactContext
	if parentContactID != "" {
		parent = dal4contactus.NewContactContext(request.TeamID, parentContactID)
		if err = tx.Get(ctx, parent.Record); err != nil {
			return contact, fmt.Errorf("failed to get parent contact with ID=[%s]: %w", parentContactID, err)
		}
	}

	teamContactus := dal4contactus.NewContactusTeamContext(request.TeamID)
	if err = tx.Get(ctx, teamContactus.Record); err != nil && !dal.IsNotFound(err) {
		return contact, fmt.Errorf("failed to get team conctacts brief record")
	}

	var contactDto models4contactus.ContactDto
	contactDto.Status = "active"
	contactDto.ParentID = parentContactID
	contactDto.WithRoles = request.WithRoles
	if request.Person != nil {
		contactDto.ContactBase = request.Person.ContactBase
		contactDto.Type = briefs4contactus.ContactTypePerson
		if contactDto.AgeGroup == "" {
			contactDto.AgeGroup = "unknown"
		}
		if contactDto.Gender == "" {
			contactDto.Gender = "unknown"
		}
		contactDto.ContactBase = request.Person.ContactBase
	} else if request.Company != nil {
		contactDto.Type = briefs4contactus.ContactTypeCompany
		contactDto.Title = request.Company.Title
		contactDto.VATNumber = request.Company.VATNumber
		contactDto.Address = request.Company.Address
	} else if request.Location != nil {
		contactDto.Type = briefs4contactus.ContactTypeLocation
		contactDto.Title = request.Location.Title
		contactDto.Address = &request.Location.Address
	} else if request.Basic != nil {
		contactDto.Type = request.Type
		contactDto.Title = request.Basic.Title
	} else {
		return contact, errors.New("contact type is not specified")
	}
	if contactDto.Address != nil {
		contactDto.CountryID = contactDto.Address.CountryID
	}
	var contactID string
	contactBrief := contactDto.ContactBrief
	if request.ContactID == "" {
		contactID, err = dbmodels.NewUniqueRandomID(teamContactus.Data.ContactIDs(), 3)
		if err != nil {
			return contact, fmt.Errorf("failed to generate new contact ContactID: %w", err)
		}
	} else {
		contactID = request.ContactID
	}
	teamContactus.Data.AddContact(contactID, &contactBrief)
	if teamContactus.Record.Exists() {
		if err = tx.Update(ctx, teamContactus.Key, []dal.Update{
			{
				Field: "contacts",
				Value: teamContactus.Data.Contacts,
			},
		}); err != nil {
			return contact, fmt.Errorf("failed to update team contact briefs: %w", err)
		}
	} else {
		if err = tx.Insert(ctx, teamContactus.Record); err != nil {
			return contact, fmt.Errorf("faield to insert team contacts brief record: %w", err)
		}
	}

	params.TeamUpdates = append(params.TeamUpdates, params.Team.Data.UpdateNumberOf("contacts", len(teamContactus.Data.Contacts)))
	contact = dal4contactus.NewContactContextWithDto(request.TeamID, contactID, &contactDto)

	//contact.Data.UserIDs = params.Team.Data.UserIDs
	if err := contact.Data.Validate(); err != nil {
		return contact, fmt.Errorf("contact record is not valid: %w", err)
	}
	if err = tx.Insert(ctx, contact.Record); err != nil {
		return contact, fmt.Errorf("failed to insert contact record: %w", err)
	}

	if parent.ID != "" {
		if err = updateParentContact(ctx, tx, contact, parent); err != nil {
			return contact, fmt.Errorf("failed to update parent contact: %w", err)
		}
	}
	return contact, err
}
