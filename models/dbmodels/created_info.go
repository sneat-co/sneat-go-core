package dbmodels

import (
	"github.com/strongo/validation"
)

type CreatedInfo struct {
	Client RemoteClientInfo `json:"client,omitempty"  firestore:"client,omitempty"`
}

func (v CreatedInfo) Validate() error {
	if err := v.Client.Validate(); err != nil {
		return validation.NewErrBadRecordFieldValue("client", err.Error())
	}
	return nil
}
