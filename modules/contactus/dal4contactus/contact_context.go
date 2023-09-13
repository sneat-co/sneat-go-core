package dal4contactus

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/sneat-co/sneat-go-core/modules/contactus/models4contactus"
)

type ContactContext = record.DataWithID[string, *models4contactus.ContactDto]

func NewContactContext(teamID, contactID string) ContactContext {
	return NewContactContextWithDto(teamID, contactID, new(models4contactus.ContactDto))
}

func NewContactContextWithDto(teamID, contactID string, dto *models4contactus.ContactDto) (contact ContactContext) {
	key := NewContactKey(teamID, contactID)
	contact.ID = contactID
	contact.FullID = teamID + ":" + contactID
	contact.Key = key
	contact.Data = dto
	contact.Record = dal.NewRecordWithData(key, dto)
	return
}

func GetContactByID(contacts []ContactContext, contactID string) (contact ContactContext, found bool) {
	for _, contact := range contacts {
		if contact.ID == contactID {
			return contact, true
		}
	}
	return contact, false
}

func ContactIDs(contacts []ContactContext) []string {
	ids := make([]string, len(contacts))
	for i, contact := range contacts {
		ids[i] = contact.ID
	}
	return ids
}
