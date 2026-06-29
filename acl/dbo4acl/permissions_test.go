package dbo4acl

import (
	"testing"
	"time"

	"github.com/sneat-co/sneat-go-core/acl/const4acl"
	"github.com/strongo/strongoapp/with"
)

func TestPermissions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		v       Permissions
		wantErr bool
	}{
		{
			name: "valid",
			v: Permissions{const4acl.PermittedToView: with.CreatedFields{
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
