package dbmodels

import "testing"

func TestPersonEmail_Validate(t *testing.T) {
	tests := []struct {
		name    string
		email   PersonEmail
		wantErr bool
	}{
		{
			name:    "empty type",
			email:   PersonEmail{Address: "test@example.com"},
			wantErr: true,
		},
		{
			name:    "unknown type",
			email:   PersonEmail{Type: "unknown", Address: "test@example.com"},
			wantErr: true,
		},
		{
			name:    "empty address",
			email:   PersonEmail{Type: "primary"},
			wantErr: true,
		},
		{
			name:    "invalid email address",
			email:   PersonEmail{Type: "primary", Address: "not-an-email"},
			wantErr: true,
		},
		{
			name:    "valid primary email",
			email:   PersonEmail{Type: "primary", Address: "test@example.com"},
			wantErr: false,
		},
		{
			name:    "valid personal email",
			email:   PersonEmail{Type: "personal", Address: "user@example.com"},
			wantErr: false,
		},
		{
			name:    "valid work email",
			email:   PersonEmail{Type: "work", Address: "work@company.com"},
			wantErr: false,
		},
		{
			name:    "valid email with note",
			email:   PersonEmail{Type: "primary", Address: "test@example.com", Note: "main"},
			wantErr: false,
		},
		{
			name:    "valid email with google provider",
			email:   PersonEmail{Type: "primary", Address: "test@example.com", AuthProvider: "google.com"},
			wantErr: false,
		},
		{
			name:    "invalid auth provider",
			email:   PersonEmail{Type: "primary", Address: "test@example.com", AuthProvider: "invalid-provider"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.email.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PersonEmail.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
