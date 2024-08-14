package dbmodels

import (
	"fmt"
	"github.com/dal-go/dalgo/dal"
)

type WithPreferredLocale struct {
	PreferredLocale string `json:"preferredLocale,omitempty" firestore:"preferredLocale,omitempty"`
}

func (v *WithPreferredLocale) GetPreferredLocale() string {
	return v.PreferredLocale
}

func (v *WithPreferredLocale) SetPreferredLocale(code5 string) (updates []dal.Update, err error) {
	if l := len(code5); l != 0 && l != 5 {
		return nil, fmt.Errorf("invalid code5: '%s'", code5)
	}
	v.PreferredLocale = code5
	updates = append(updates, dal.Update{Field: "preferredLocale", Value: code5})
	return updates, err
}
