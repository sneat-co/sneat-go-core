package dbmodels

import "testing"

// Simple type for testing
type testData struct {
	Value string
}

func (t testData) Validate() error {
	return nil
}

func TestDtoWithID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		dto     *DtoWithID[testData]
		wantErr bool
	}{
		{
			name:    "nil dto",
			dto:     nil,
			wantErr: false,
		},
		{
			name: "empty ID",
			dto: &DtoWithID[testData]{
				ID:   "",
				Data: testData{Value: "test"},
			},
			wantErr: true,
		},
		{
			name: "valid dto",
			dto: &DtoWithID[testData]{
				ID:   "id123",
				Data: testData{Value: "test"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dto.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DtoWithID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
