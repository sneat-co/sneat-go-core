package dbmodels

import "testing"

func TestWithUserID_GetUserID(t *testing.T) {
	v := WithUserID{UserID: "user123"}
	if got := v.GetUserID(); got != "user123" {
		t.Errorf("GetUserID() = %v, want %v", got, "user123")
	}
}

func TestWithUserID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{
			name:    "empty user ID",
			userID:  "",
			wantErr: false,
		},
		{
			name:    "valid user ID",
			userID:  "user123",
			wantErr: false,
		},
		{
			name:    "user ID with leading space",
			userID:  " user123",
			wantErr: true,
		},
		{
			name:    "user ID with trailing space",
			userID:  "user123 ",
			wantErr: true,
		},
		{
			name:    "user ID with leading and trailing spaces",
			userID:  " user123 ",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := WithUserID{UserID: tt.userID}
			err := v.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("WithUserID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
