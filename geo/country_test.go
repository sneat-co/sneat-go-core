package geo

import "testing"

func TestIsValidCountryAlpha2(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if !IsValidCountryAlpha2("US") {
			t.Error("expected true")
		}
	})
	t.Run("invalid", func(t *testing.T) {
		if IsValidCountryAlpha2("USA") {
			t.Error("expected false")
		}
	})
}
