package dbmodels

import (
	"github.com/crediterra/money"
	"testing"
)

func TestWithPrimaryCurrency_SetPrimaryCurrency(t *testing.T) {
	v := &WithPrimaryCurrency{}
	const code3 money.CurrencyCode = "USD"
	updates, err := v.SetPrimaryCurrency(code3)
	if err != nil {
		t.Error(err)
	}
	if len(updates) != 1 {
		t.Error("Expected 1 update, got", len(updates))
	}
	if fieldName := updates[0].FieldName(); fieldName != "primaryCurrency" {
		t.Error("Expected 'primaryCurrency', got", fieldName)
	}
	if v := updates[0].Value(); v != code3 {
		t.Errorf("Expected '%s', got '%s'", code3, v)
	}
}
