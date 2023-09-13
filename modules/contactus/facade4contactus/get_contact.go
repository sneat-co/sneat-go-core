package facade4contactus

import (
	"context"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/modules/contactus/dal4contactus"
)

func GetContactByID(ctx context.Context, tx dal.ReadSession, teamID, contactID string) (contact dal4contactus.ContactContext, err error) {
	contact = dal4contactus.NewContactContext(teamID, contactID)
	return contact, tx.Get(ctx, contact.Record)
}
