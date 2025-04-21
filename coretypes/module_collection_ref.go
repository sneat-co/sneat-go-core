package coretypes

import "github.com/strongo/validation"

type ModuleCollectionRef struct {
	ModuleID   ModuleID `json:"moduleID" firestore:"moduleID"`
	Collection string   `json:"collection" firestore:"collection"`
}

func (v *ModuleCollectionRef) Validate() error {
	if v.ModuleID == "" {
		return validation.NewErrRequestIsMissingRequiredField("moduleID")
	}
	if v.Collection == "" {
		return validation.NewErrRequestIsMissingRequiredField("collection")
	}
	return nil

}
