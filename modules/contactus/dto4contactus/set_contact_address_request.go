package dto4contactus

import (
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/strongo/validation"
)

// SetContactAddressRequest request to set contact address
type SetContactAddressRequest struct {
	ContactRequest
	Address dbmodels.Address `json:"address"`
}

// Validate returns error if request is invalid
func (v SetContactAddressRequest) Validate() error {
	if err := v.ContactRequest.Validate(); err != nil {
		return err
	}
	if err := v.Address.Validate(); err != nil {
		return validation.NewErrBadRequestFieldValue("address", err.Error())
	}
	return nil
}
