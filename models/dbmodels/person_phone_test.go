package dbmodels

import "testing"

func TestPersonPhone_Validate(t *testing.T) {
	tests := []struct {
		name    string
		phone   PersonPhone
		wantErr bool
	}{
		{
			name:    "empty type",
			phone:   PersonPhone{Number: "123456789"},
			wantErr: true,
		},
		{
			name:    "empty number",
			phone:   PersonPhone{Type: "mobile"},
			wantErr: true,
		},
		{
			name:    "valid phone",
			phone:   PersonPhone{Type: "mobile", Number: "123456789"},
			wantErr: false,
		},
		{
			name:    "valid phone with note",
			phone:   PersonPhone{Type: "home", Number: "987654321", Note: "primary"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.phone.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PersonPhone.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
