package dbmodels

import (
	"github.com/crediterra/money"
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/validation"
)

type WithLastCurrencies struct {
	LastCurrencies []money.CurrencyCode `json:"lastCurrencies,omitempty" dalgo:"lastCurrencies,omitempty" firestore:"lastCurrencies,omitempty"`
}

func (v *WithLastCurrencies) GetLastCurrencies() (lastCurrencies []money.CurrencyCode) {
	lastCurrencies = make([]money.CurrencyCode, len(v.LastCurrencies))
	copy(lastCurrencies, v.LastCurrencies)
	return
}

func (v *WithLastCurrencies) SetLastCurrency(currencyCode money.CurrencyCode) (updates []dal.Update, err error) {
	if !money.IsKnownCurrency(currencyCode) {
		return nil, validation.NewErrBadRecordFieldValue("currencyCode", "unknown currency")
	}
	if len(v.LastCurrencies) > 0 && v.LastCurrencies[0] == currencyCode {
		return nil, nil
	}
	for i, c := range v.LastCurrencies {
		if c == currencyCode {
			if i > 0 {
				for j := i - 1; j >= 0; j-- {
					v.LastCurrencies[j+1] = v.LastCurrencies[j]
				}
				v.LastCurrencies[0] = currencyCode
			}
			return
		}
	}
	v.LastCurrencies = append([]money.CurrencyCode{currencyCode}, v.LastCurrencies...)
	if len(v.LastCurrencies) > 10 {
		v.LastCurrencies = v.LastCurrencies[:10]
	}
	updates = []dal.Update{{Field: "lastCurrencies", Value: v.LastCurrencies}}
	return
}
