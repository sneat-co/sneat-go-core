package dto4auth

import (
	"fmt"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/strongo/strongoapp/appuser"
	"github.com/strongo/strongoapp/person"
	"github.com/strongo/validation"
)

// DataToCreateUser is NOT a DTO object - do not user at transport layer!
type DataToCreateUser struct {
	AuthAccount     appuser.AccountKey
	Email           string
	EmailIsVerified bool
	// AgeGroup is optional; when set it must be one of dbmodels.AgeGroups.
	AgeGroup string
	// Gender is optional; when set it must be one of dbmodels.Genders.
	Gender       dbmodels.Gender
	IanaTimezone string
	LanguageCode string
	Names        person.NameFields
	PhotoURL     string
	RemoteClient dbmodels.RemoteClientInfo
}

func (v DataToCreateUser) Validate() error {
	if err := v.AuthAccount.Validate(); err != nil {
		return validation.NewErrBadRequestFieldValue("authAccount", "mismatch with account provider")
	}
	if err := v.RemoteClient.Validate(); err != nil {
		return fmt.Errorf("invalid remote client: %w", err)
	}
	// The value is not echoed: it is untrusted client text.
	if !dbmodels.IsKnownGenderOrEmpty(v.Gender) {
		return validation.NewErrBadRequestFieldValue("gender", "unknown gender")
	}
	if !dbmodels.IsKnownAgeGroupOrEmpty(v.AgeGroup) {
		return validation.NewErrBadRequestFieldValue("ageGroup", "unknown age group")
	}
	return nil
}
