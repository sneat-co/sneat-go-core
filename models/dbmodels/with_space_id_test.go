package dbmodels

import (
	"github.com/sneat-co/sneat-go-core/coretypes"
	"testing"
)

func TestWithSpaceID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		spaceID coretypes.SpaceID
		wantErr bool
	}{
		{
			name:    "empty space ID",
			spaceID: "",
			wantErr: true,
		},
		{
			name:    "whitespace space ID",
			spaceID: "  ",
			wantErr: true,
		},
		{
			name:    "valid space ID",
			spaceID: "space123",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := WithSpaceID{SpaceID: tt.spaceID}
			err := v.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("WithSpaceID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
