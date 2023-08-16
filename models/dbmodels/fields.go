package dbmodels

import (
	"github.com/crediterra/money"
	"github.com/sneat-co/sneat-go-core/validate"
	"github.com/strongo/validation"
)

type ByLang = map[string]string

type WithCustomFields struct {
	FieldsStr    map[string]ByLang       `json:"fieldsStr,omitempty" firestore:"fieldsStr,omitempty"`
	FieldsInt    map[string]int          `json:"fieldsInt,omitempty" firestore:"fieldsInt,omitempty"`
	FieldsDate   map[string]string       `json:"fieldsDate,omitempty" firestore:"fieldsDate,omitempty"`
	FieldsAmount map[string]money.Amount `json:"fieldsAmount,omitempty" firestore:"fieldsAmount,omitempty"`
}

func (v *WithCustomFields) Validate() error {
	for n, v := range v.FieldsDate {
		if _, err := validate.DateString(v); err != nil {
			return validation.NewErrBadRecordFieldValue("fieldsDate."+n, err.Error())
		}
	}
	for n, v := range v.FieldsAmount {
		if err := v.Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue("fieldsAmount."+n, err.Error())
		}
	}
	return nil
}
