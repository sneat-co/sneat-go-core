package dbmodels

import (
	"github.com/crediterra/money"
	"testing"
)

func TestWithLastCurrencies_SetLastCurrency(t *testing.T) {
	dbo := WithLastCurrencies{}
	updates, err := dbo.SetLastCurrency("EUR")
	if err != nil {
		t.Fatalf("Failed to set last currency: %v", err)
	}
	if len(updates) != 1 {
		t.Errorf("Expected 1 update, got: %d", len(updates))
	}
	if len(dbo.LastCurrencies) != 1 {
		t.Errorf("Expected 1 value in LastCurrencies, got: %d", len(dbo.LastCurrencies))
	}
	updates, err = dbo.SetLastCurrency("USD")
	if err != nil {
		t.Fatalf("Failed to set last currency: %v", err)
	}
	if len(updates) != 1 {
		t.Errorf("Expected 1 update, got: %d", len(updates))
	}
	if len(dbo.LastCurrencies) != 2 {
		t.Errorf("Expected 2 values in LastCurrencies, got: %d", len(dbo.LastCurrencies))
	}
	if dbo.LastCurrencies[0] != "USD" {
		t.Errorf("First currency should be USD, got: %v", dbo.LastCurrencies[0])
	}
	if dbo.LastCurrencies[1] != "EUR" {
		t.Errorf("Second currency should be EUR, got: %v", dbo.LastCurrencies[1])
	}

	updates, err = dbo.SetLastCurrency("EUR")
	if err != nil {
		t.Fatalf("Failed to set last currency: %v", err)
	}
	if len(updates) != 0 {
		t.Errorf("Expected 1 update, got: %d", len(updates))
	}
	if len(dbo.LastCurrencies) != 2 {
		t.Errorf("Expected 2 values in LastCurrencies, got: %d", len(dbo.LastCurrencies))
	}
	if dbo.LastCurrencies[0] != "EUR" {
		t.Errorf("Second currency should be EUR, got: %v", dbo.LastCurrencies[0])
	}
	if dbo.LastCurrencies[1] != "USD" {
		t.Errorf("First currency should be USD, got: %v", dbo.LastCurrencies[1])
	}
}

func TestWithLastCurrencies_GetLastCurrencies(t *testing.T) {
	v := &WithLastCurrencies{
		LastCurrencies: []money.CurrencyCode{"USD", "EUR"},
	}
	lastCurrencies := v.GetLastCurrencies()
	if len(lastCurrencies) != 2 {
		t.Errorf("Expected 2 currencies, got: %d", len(lastCurrencies))
	}
	if lastCurrencies[0] != "USD" {
		t.Errorf("Expected USD, got: %v", lastCurrencies[0])
	}
	if lastCurrencies[1] != "EUR" {
		t.Errorf("Expected EUR, got: %v", lastCurrencies[1])
	}
	lastCurrencies[0] = "RUB"
	if v.LastCurrencies[0] != "USD" {
		t.Errorf("Expected USD, got: %v", v.LastCurrencies[0])
	}
}
