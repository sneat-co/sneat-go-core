package dbmodels

import "testing"

func TestValidateTitle(t *testing.T) {
	if err := ValidateTitle(""); err == nil {
		t.Error("expected an error got nil")
	}
	if err := ValidateTitle(" "); err == nil {
		t.Error("expected an error got nil")
	}
	if err := ValidateTitle("\t"); err == nil {
		t.Error("expected an error got nil")
	}
}
