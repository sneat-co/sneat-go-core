package dal4contactus

import (
	"github.com/dal-go/dalgo/record"
	"github.com/sneat-co/sneat-go-core/modules/contactus/models4contactus"
	"github.com/sneat-co/sneat-go-core/modules/teamus/dal4teamus"
)

type ContactusTeamContext = record.DataWithID[string, *models4contactus.ContactusTeamDto]

func NewContactusTeamContext(teamID string) ContactusTeamContext {
	return NewContactusTeamContextWithData(teamID, new(models4contactus.ContactusTeamDto))
}

func NewContactusTeamContextWithData(teamID string, data *models4contactus.ContactusTeamDto) ContactusTeamContext {
	key := dal4teamus.NewTeamModuleKey(teamID, "contactus")
	return record.NewDataWithID(teamID, key, data)
}
