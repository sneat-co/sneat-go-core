package dbo4sharing

import (
	"github.com/sneat-co/sneat-go-core/sharing/const4sharing"
	"github.com/strongo/strongoapp/with"
	"testing"
	"time"
)

func TestOfferDbo_Validate(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		offer   OfferDbo
		wantErr bool
	}{
		{
			name: "valid offer",
			offer: OfferDbo{
				CreatedFields: with.CreatedFields{
					CreatedAtField: with.CreatedAtField{CreatedAt: now},
					CreatedByField: with.CreatedByField{CreatedBy: "user1"},
				},
				Permissions: []const4sharing.Permission{"read"},
			},
			wantErr: false,
		},
		{
			name: "empty permissions",
			offer: OfferDbo{
				CreatedFields: with.CreatedFields{
					CreatedAtField: with.CreatedAtField{CreatedAt: now},
					CreatedByField: with.CreatedByField{CreatedBy: "user1"},
				},
				Permissions: []const4sharing.Permission{},
			},
			wantErr: true,
		},
		{
			name: "nil permissions",
			offer: OfferDbo{
				CreatedFields: with.CreatedFields{
					CreatedAtField: with.CreatedAtField{CreatedAt: now},
					CreatedByField: with.CreatedByField{CreatedBy: "user1"},
				},
				Permissions: nil,
			},
			wantErr: true,
		},
		{
			name: "empty permission string",
			offer: OfferDbo{
				CreatedFields: with.CreatedFields{
					CreatedAtField: with.CreatedAtField{CreatedAt: now},
					CreatedByField: with.CreatedByField{CreatedBy: "user1"},
				},
				Permissions: []const4sharing.Permission{"read", ""},
			},
			wantErr: true,
		},
		{
			name: "multiple valid permissions",
			offer: OfferDbo{
				CreatedFields: with.CreatedFields{
					CreatedAtField: with.CreatedAtField{CreatedAt: now},
					CreatedByField: with.CreatedByField{CreatedBy: "user1"},
				},
				Permissions: []const4sharing.Permission{"read", "write"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.offer.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("OfferDbo.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
