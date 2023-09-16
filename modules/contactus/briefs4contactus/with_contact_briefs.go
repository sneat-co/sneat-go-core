package briefs4contactus

import (
	"fmt"
	"github.com/sneat-co/sneat-go-core"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/strongo/validation"
)

// WithContactBriefs is a base struct for DTOs that have contacts
// TODO: Document how it is different from WithContactsBase or merge them
type WithContactBriefs[
	K string | dbmodels.TeamItemID,
	T interface {
		core.Validatable
		Equal(v T) bool
	},
] struct {
	Contacts map[K]T `json:"contacts,omitempty" firestore:"contacts,omitempty"`
}

// WithContactsBase is a base struct for DTOs that represent a short version of a contact
// TODO: Document how it is different from WithContactBriefs or merge them
type WithContactsBase[
	K string | dbmodels.TeamItemID,
	T interface {
		dbmodels.UserIDGetter
		dbmodels.RelatedAs
		HasRole(role string) bool
		Equal(v T) bool
	}] struct {
	WithContactBriefs[K, T]
	dbmodels.WithUserIDs
}

func (v WithContactsBase[K, T]) Validate() error {
	for id, contact := range v.Contacts {
		if err := contact.Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue("contacts."+string(id), err.Error())
		}
		if userID := contact.GetUserID(); userID == "" {
			if !v.HasUserID(userID) {
				return validation.NewErrBadRecordFieldValue(
					fmt.Sprintf("contacts.%s.userID", id),
					fmt.Sprintf("%s not added to userIDs", userID))
			}
		}
	}
	return nil
}

func (v WithContactsBase[K, T]) GetContactBriefByUserID(userID string) (id K, contactBrief T) {
	for k, c := range v.Contacts {
		if c.GetUserID() == userID {
			return k, c
		}
	}
	return id, contactBrief
}

func (v WithContactsBase[K, T]) GetContactBriefsByRoles(roles ...string) map[K]T {
	result := make(map[K]T)
	for id, c := range v.Contacts {
		for _, role := range roles {
			if c.HasRole(role) {
				result[id] = c
				break
			}
		}
	}
	return result
}
