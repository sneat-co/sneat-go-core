package dbmodels

import (
	"fmt"
)

// WithCountryIDs defines a record with a Country IDs
type WithCountryIDs struct {
	CountryIDs []string `json:"countryIDs,omitempty" firestore:"countryIDs,omitempty"`
}

func (v WithCountryIDs) Validate() error {
	for i, countryID := range v.CountryIDs {
		if err := ValidateRequiredCountryID(fmt.Sprintf("countryIDs[%v]", i), countryID); err != nil {
			return err
		}
	}
	return nil
}
