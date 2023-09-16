package briefs4contactus

import (
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/validate"
	"github.com/strongo/slice"
	"github.com/strongo/validation"
)

type contactBrief interface {
	dbmodels.UserIDGetter
	dbmodels.RelatedAs
}

type WithSingleTeamContactsWithoutContactIDs[
	T interface {
		contactBrief
		HasRole(role string) bool
		Equal(v T) bool
	},
] struct {
	WithContactsBase[string, T]
}

func (v *WithSingleTeamContactsWithoutContactIDs[T]) Validate() error {
	for id, brief := range v.Contacts {
		if err := validate.RecordID(id); err != nil {
			return validation.NewErrBadRecordFieldValue("contacts",
				fmt.Sprintf("invalid contact ID=%s: %v", id, err))
		}
		if err := brief.Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue("contacts."+id, err.Error())
		}
	}
	return nil
}

func (v *WithSingleTeamContactsWithoutContactIDs[T]) ContactIDs() (contactIDs []string) {
	contactIDs = make([]string, 0, len(v.Contacts))
	for id := range v.Contacts {
		contactIDs = append(contactIDs, id)
	}
	return
}

func (v *WithSingleTeamContactsWithoutContactIDs[T]) HasContact(contactID string) bool {
	_, ok := v.Contacts[contactID]
	return ok
}

func (v *WithSingleTeamContactsWithoutContactIDs[T]) AddContact(contactID string, contact T) dal.Update {
	if v.Contacts == nil {
		v.Contacts = make(map[string]T)
	}
	v.Contacts[contactID] = contact
	return dal.Update{
		Field: "contacts." + contactID,
		Value: contact,
	}
}

func (v *WithSingleTeamContactsWithoutContactIDs[T]) RemoveContact(contactID string) dal.Update {
	delete(v.Contacts, contactID)
	return dal.Update{
		Field: "contacts." + contactID,
		Value: dal.DeleteField,
	}
}

// WithMultiTeamContacts mixin that adds WithMultiTeamContactIDs.ContactIDs & Contacts fields
type WithMultiTeamContacts[
	T interface {
		contactBrief
		HasRole(role string) bool
		Equal(v T) bool
	},
] struct {
	WithMultiTeamContactIDs
	WithContactsBase[dbmodels.TeamItemID, T]
}

// Validate returns error if not valid
func (v *WithMultiTeamContacts[T]) Validate() error {
	if err := v.WithMultiTeamContactIDs.Validate(); err != nil {
		return nil
	}
	return dbmodels.ValidateWithIdsAndBriefs("contactIDs", "contacts", v.ContactIDs, v.Contacts)
}

func (v *WithMultiTeamContacts[T]) Updates(contactIDs ...dbmodels.TeamItemID) (updates []dal.Update) {
	updates = append(updates,
		dal.Update{
			Field: "contactIDs",
			Value: v.ContactIDs,
		},
	)
	if len(contactIDs) == 0 {
		updates = append(updates, dal.Update{
			Field: "contacts",
			Value: v.Contacts,
		})
	} else {
		for _, id := range contactIDs {
			updates = append(updates, dal.Update{
				Field: "contacts." + string(id),
				Value: v.Contacts[id],
			})
		}
	}
	return
}

// SetContactBrief sets contactBrief brief by ID
func (v *WithMultiTeamContacts[T]) SetContactBrief(teamID, contactID string, contactBrief T) (updates []dal.Update) {
	id := dbmodels.NewTeamItemID(teamID, contactID)
	if !slice.Contains(v.ContactIDs, id) {
		v.ContactIDs = append(v.ContactIDs, id)
		updates = append(updates, dal.Update{
			Field: "contactIDs",
			Value: v.ContactIDs,
		})
	}
	if currentBrief, ok := v.Contacts[id]; !ok || !currentBrief.Equal(contactBrief) {
		v.Contacts[id] = contactBrief
		updates = append(updates, dal.Update{
			Field: "contacts" + string(id),
			Value: contactBrief,
		})
	}
	return
}

// ParentContactBrief returns parent contactBrief brief
func (v *WithMultiTeamContacts[T]) ParentContactBrief() (i int, id dbmodels.TeamItemID, brief T) {
	for i, id := range v.ContactIDs {
		brief := v.Contacts[id]
		if brief.GetRelatedAs() == "parent" {
			return i, id, brief
		}
	}
	return -1, "", brief
}

// GetContactBriefByID returns contactBrief brief by ID
func (v *WithMultiTeamContacts[T]) GetContactBriefByID(teamID, contactID string) (i int, brief T) {
	id := dbmodels.NewTeamItemID(teamID, contactID)
	if brief, ok := v.Contacts[id]; !ok {
		return -1, brief
	}
	return slice.Index(v.ContactIDs, id), brief
}

// GetContactBriefByUserID returns contactBrief brief by user ID
func (v *WithMultiTeamContacts[T]) GetContactBriefByUserID(userID string) (id dbmodels.TeamItemID, t T) {
	for cID, c := range v.Contacts {
		if c.GetUserID() == userID {
			return cID, c
		}
	}
	return
}

func (v *WithMultiTeamContacts[T]) AddContact(teamID, contactID string, c T) (updates []dal.Update) {
	id := dbmodels.NewTeamItemID(teamID, contactID)
	if !slice.Contains(v.ContactIDs, id) {
		v.ContactIDs = append(v.ContactIDs, id)
		updates = append(updates, dal.Update{
			Field: "contactIDs",
			Value: v.ContactIDs,
		})
	}
	if _, ok := v.Contacts[id]; !ok {
		updates = append(updates, dal.Update{
			Field: "contacts." + string(id),
			Value: c,
		})
	}
	if v.Contacts == nil {
		v.Contacts = make(map[dbmodels.TeamItemID]T)
	}
	v.Contacts[id] = c
	return
}
