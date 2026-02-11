package dbmodels

import "testing"

func TestIsKnownGender(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"male", GenderMale, true},
		{"female", GenderFemale, true},
		{"nonbinary", GenderNonbinary, true},
		{"unknown", GenderUnknown, true},
		{"other", GenderTypeOther, true},
		{"undisclosed", GenderTypeUndisclosed, true},
		{"invalid value", "invalid", false},
		{"empty string", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsKnownGender(tt.value); got != tt.want {
				t.Errorf("IsKnownGender(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestIsKnownGenderOrEmpty(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"empty string", "", true},
		{"male", GenderMale, true},
		{"female", GenderFemale, true},
		{"invalid value", "invalid", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsKnownGenderOrEmpty(tt.value); got != tt.want {
				t.Errorf("IsKnownGenderOrEmpty(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestValidateGender(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		required bool
		wantErr  bool
	}{
		{"valid male", GenderMale, false, false},
		{"valid female", GenderFemale, false, false},
		{"empty not required", "", false, false},
		{"empty required", "", true, true},
		{"whitespace required", "  ", true, true},
		{"invalid value", "invalid", false, true},
		{"invalid value required", "invalid", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGender(tt.value, tt.required)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGender(%v, %v) error = %v, wantErr %v", tt.value, tt.required, err, tt.wantErr)
			}
		})
	}
}
