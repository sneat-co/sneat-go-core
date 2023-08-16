package dbmodels

import (
	"github.com/strongo/validation"
	"time"
)

type CreatedInfo struct {
	At time.Time `json:"at,omitempty"  firestore:"at,omitempty"`
	//
	Client RemoteClientInfo `json:"client,omitempty"  firestore:"client,omitempty"`
}

func (v CreatedInfo) Validate() error {
	if v.At.IsZero() {
		return validation.NewErrRecordIsMissingRequiredField("at")
	}
	if err := v.Client.Validate(); err != nil {
		return validation.NewErrBadRecordFieldValue("client", err.Error())
	}
	return nil
}
