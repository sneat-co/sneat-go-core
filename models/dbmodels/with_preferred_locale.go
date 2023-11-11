package dbmodels

import "fmt"

type WithPreferredLocale struct {
	PreferredLocale string `json:"preferredLocale,omitempty" firestore:"preferredLocale,omitempty"`
}

func (v *WithPreferredLocale) GetPreferredLocale() string {
	return v.PreferredLocale
}

func (v *WithPreferredLocale) SetPreferredLocale(code5 string) {
	if l := len(code5); l != 0 && l != 5 {
		panic(fmt.Sprintf("invalid code5: '%s'", code5))
	}
	v.PreferredLocale = code5
}
