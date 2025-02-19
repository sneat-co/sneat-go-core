package dbmodels

import (
	"github.com/crediterra/money"
	"github.com/dal-go/dalgo/update"
	"github.com/strongo/validation"
)

type WithPrimaryCurrency struct {
	PrimaryCurrency money.CurrencyCode `json:"primaryCurrency,omitempty" firestore:"primaryCurrency,omitempty"`
}

func (v *WithPrimaryCurrency) Validate() error {
	if v.PrimaryCurrency == "" {
		return nil
	}
	if !money.IsKnownCurrency(v.PrimaryCurrency) {
		return validation.NewErrBadRecordFieldValue("primaryCurrency", "unknown currency code: "+string(v.PrimaryCurrency))
	}
	return nil
}

func (v *WithPrimaryCurrency) SetPrimaryCurrency(currencyCode money.CurrencyCode) (updates []update.Update, err error) {
	v.PrimaryCurrency = currencyCode
	updates = append(updates, update.ByFieldName("primaryCurrency", v.PrimaryCurrency))
	return
}
