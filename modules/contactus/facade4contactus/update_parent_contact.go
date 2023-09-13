package facade4contactus

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/modules/contactus/briefs4contactus"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dal4contactus"
	"log"
)

func updateParentContact(
	ctx context.Context,
	tx dal.ReadwriteTransaction,
	contact, parent dal4contactus.ContactContext,
) error {
	log.Printf("updateParentContact(contact=%v, parentID=%v)", contact.ID, parent.ID)
	contactBrief := &briefs4contactus.ContactBrief{
		Type:   contact.Data.Type,
		Gender: contact.Data.Gender,
		Name:   contact.Data.Name,
	}
	contactBrief.RelatedAs = RelatedAsChild
	updates := parent.Data.SetContactBrief(contact.Key.Parent().ID.(string), contact.ID, contactBrief)
	if err := parent.Data.Validate(); err != nil {
		return fmt.Errorf("parent contact DTO validation error: %w", err)
	}
	if err := updateContact(ctx, tx, parent.Key, updates); err != nil {
		return fmt.Errorf("failed to update parent contact record: %w", err)
	}
	log.Printf("updateParentContact(contact=%v, parentID=%v) - success!", contact.ID, parent.ID)
	return nil
}
