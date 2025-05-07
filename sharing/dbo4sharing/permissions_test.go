package dbo4sharing

import (
	"github.com/sneat-co/sneat-go-core/sharing/const4sharing"
	"github.com/strongo/strongoapp/with"
	"testing"
	"time"
)

func TestPermissions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		v       Permissions
		wantErr bool
	}{
		{
			name: "valid",
			v: Permissions{const4sharing.PermittedToView: with.CreatedFields{
				CreatedAtField: with.CreatedAtField{
					CreatedAt: time.Now(),
				},
				CreatedByField: with.CreatedByField{
					CreatedBy: "user1",
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
