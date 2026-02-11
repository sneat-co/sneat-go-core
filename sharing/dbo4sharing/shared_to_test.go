package dbo4sharing

import (
	"github.com/sneat-co/sneat-go-core/sharing/const4sharing"
	"github.com/strongo/strongoapp/with"
	"testing"
	"time"
)

func TestTo_Validate(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		to      To
		wantErr bool
	}{
		{
			name:    "empty To",
			to:      To{},
			wantErr: false,
		},
		{
			name: "valid spaces",
			to: To{
				Spaces: map[string]Shared{
					"space1": {
						ID: "item1",
						Permissions: Permissions{
							const4sharing.PermittedToView: with.CreatedFields{
								CreatedAtField: with.CreatedAtField{CreatedAt: now},
								CreatedByField: with.CreatedByField{CreatedBy: "user1"},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid space - empty ID",
			to: To{
				Spaces: map[string]Shared{
					"space1": {
						ID: "",
						Permissions: Permissions{
							const4sharing.PermittedToView: with.CreatedFields{
								CreatedAtField: with.CreatedAtField{CreatedAt: now},
								CreatedByField: with.CreatedByField{CreatedBy: "user1"},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid users",
			to: To{
				Users: map[string]Permissions{
					"user1": {
						const4sharing.PermittedToView: with.CreatedFields{
							CreatedAtField: with.CreatedAtField{CreatedAt: now},
							CreatedByField: with.CreatedByField{CreatedBy: "user1"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid users - empty permissions",
			to: To{
				Users: map[string]Permissions{
					"user1": {},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.to.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("To.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestShared_Validate(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		shared  Shared
		wantErr bool
	}{
		{
			name: "valid shared",
			shared: Shared{
				ID: "item1",
				Permissions: Permissions{
					const4sharing.PermittedToView: with.CreatedFields{
						CreatedAtField: with.CreatedAtField{CreatedAt: now},
						CreatedByField: with.CreatedByField{CreatedBy: "user1"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			shared: Shared{
				ID: "",
				Permissions: Permissions{
					const4sharing.PermittedToView: with.CreatedFields{
						CreatedAtField: with.CreatedAtField{CreatedAt: now},
						CreatedByField: with.CreatedByField{CreatedBy: "user1"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty permissions",
			shared: Shared{
				ID:          "item1",
				Permissions: Permissions{},
			},
			wantErr: true,
		},
		{
			name: "nil permissions",
			shared: Shared{
				ID:          "item1",
				Permissions: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.shared.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Shared.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
