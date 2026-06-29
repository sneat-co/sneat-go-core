package dbo4sharing

import (
	"testing"
	"time"

	"github.com/sneat-co/sneat-go-core/acl/const4acl"
	"github.com/sneat-co/sneat-go-core/acl/dbo4acl"
	"github.com/strongo/strongoapp/with"
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
						Permissions: dbo4acl.Permissions{
							const4acl.PermittedToView: with.CreatedFields{
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
						Permissions: dbo4acl.Permissions{
							const4acl.PermittedToView: with.CreatedFields{
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
				ACL: dbo4acl.ACL{
					Users: map[string]dbo4acl.Permissions{
						"user1": {
							const4acl.PermittedToView: with.CreatedFields{
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
			name: "invalid users - empty permissions",
			to: To{
				ACL: dbo4acl.ACL{
					Users: map[string]dbo4acl.Permissions{
						"user1": {},
					},
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
				Permissions: dbo4acl.Permissions{
					const4acl.PermittedToView: with.CreatedFields{
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
				Permissions: dbo4acl.Permissions{
					const4acl.PermittedToView: with.CreatedFields{
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
				Permissions: dbo4acl.Permissions{},
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
