package dbmodels

import "testing"

func TestWithPreferredLocale_GetPreferredLocale(t *testing.T) {
	v := &WithPreferredLocale{PreferredLocale: "en"}
	if v.GetPreferredLocale() != "en" {
		t.Error("Expected 'en', got", v.GetPreferredLocale())
	}
}

func TestWithPreferredLocale_SetPreferredLocale(t *testing.T) {
	v := &WithPreferredLocale{}
	const code5 = "en-US"
	updates, err := v.SetPreferredLocale(code5)
	if err != nil {
		t.Error(err)
	}
	if len(updates) != 1 {
		t.Error("Expected 1 update, got", len(updates))
	}
	if updates[0].Field != "preferredLocale" {
		t.Error("Expected 'preferredLocale', got", updates[0].Field)
	}
	if updates[0].Value != code5 {
		t.Errorf("Expected '%s', got '%s'", code5, updates[0].Value)
	}
}
