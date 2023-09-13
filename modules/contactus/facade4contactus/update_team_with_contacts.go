package facade4contactus

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/modules/teamus/models4teamus"
)

func updateTeamDtoWithNumberOfContact(numberOfContacts int) (update dal.Update) {
	var value interface{}
	if numberOfContacts == 0 {
		value = dal.DeleteField
	} else {
		value = numberOfContacts
	}
	return dal.Update{
		Field: models4teamus.NumberOfUpdateField("contacts"),
		Value: value,
	}
}
