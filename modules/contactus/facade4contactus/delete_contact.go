package facade4contactus

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dal4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dto4contactus"
	"github.com/strongo/validation"
)

// UpdateContact updates team contact
func UpdateContact(ctx context.Context, tx dal.ReadwriteTransaction, contact dal4contactus.ContactContext, updates []dal.Update) (err error) {
	if err = contact.Data.Validate(); err != nil {
		return
	}
	return tx.Set(ctx, contact.Record)
}

// DeleteContact deletes team contact
func DeleteContact(ctx context.Context, userContext facade.User, request dto4contactus.ContactRequest) (err error) {
	if err = request.Validate(); err != nil {
		return
	}

	return dal4contactus.RunContactusTeamWorker(ctx, userContext, request.TeamRequest,
		func(ctx context.Context, tx dal.ReadwriteTransaction, params *dal4contactus.ContactusTeamWorkerParams) (err error) {
			return deleteContactTxWorker(ctx, tx, params, request.ContactID)
		},
	)
}

func deleteContactTxWorker(
	ctx context.Context, tx dal.ReadwriteTransaction, params *dal4contactus.ContactusTeamWorkerParams,
	contactID string,
) (err error) {
	if contactID == params.Team.ID {
		return validation.NewErrBadRequestFieldValue("contactID", "cannot delete contact that represents team/company itself")
	}
	contact := dal4contactus.NewContactContext(params.Team.ID, contactID)
	if err = tx.Get(ctx, contact.Record); err != nil {
		return fmt.Errorf("failed to get contact: %w", err)
	}
	if err = tx.Get(ctx, params.ContactusTeam.Record); err != nil && !dal.IsNotFound(err) {
		return fmt.Errorf("failed to get team contacts brief: %w", err)
	}

	var relatedContacts []dal4contactus.ContactContext
	relatedContacts, err = GetRelatedContacts(ctx, tx, params.Team.ID, RelatedAsChild, 0, -1, []dal4contactus.ContactContext{contact})
	if err != nil {
		return fmt.Errorf("failed to get related contacts: %w", err)
	}
	params.ContactusTeamUpdates = append(params.ContactusTeamUpdates,
		params.ContactusTeam.Data.RemoveContact(contactID))

	if err := params.Team.Data.Validate(); err != nil {
		return err
	}

	params.TeamUpdates = append(params.TeamUpdates, updateTeamDtoWithNumberOfContact(len(params.ContactusTeam.Data.Contacts)))

	contactKeysToDelete := make([]*dal.Key, 0, len(relatedContacts)+1)
	contactKeysToDelete = append(contactKeysToDelete, contact.Key)
	for _, relatedContact := range relatedContacts {
		contactKeysToDelete = append(contactKeysToDelete, relatedContact.Key)
	}
	if err = tx.DeleteMulti(ctx, contactKeysToDelete); err != nil {
		return fmt.Errorf("failed to delete contacts: %w", err)
	}
	return nil
}
