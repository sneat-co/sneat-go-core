package dbo4acl

import (
	"testing"
	"time"

	"github.com/sneat-co/sneat-go-core/acl/const4acl"
	"github.com/strongo/strongoapp/with"
)

func grant(by string) with.CreatedFields {
	return with.CreatedFields{
		CreatedAtField: with.CreatedAtField{CreatedAt: time.Now()},
		CreatedByField: with.CreatedByField{CreatedBy: by},
	}
}

func TestACL_UserCan(t *testing.T) {
	acl := ACL{Users: map[string]Permissions{
		"editor": {const4acl.PermittedToEdit: grant("owner")},
		"viewer": {const4acl.PermittedToView: grant("owner")},
	}}
	tests := []struct {
		name   string
		acl    ACL
		userID string
		perm   const4acl.Permission
		want   bool
	}{
		{"granted edit", acl, "editor", const4acl.PermittedToEdit, true},
		{"granted view is not edit", acl, "viewer", const4acl.PermittedToEdit, false},
		{"granted view", acl, "viewer", const4acl.PermittedToView, true},
		{"unknown user denied", acl, "stranger", const4acl.PermittedToEdit, false},
		{"empty user denied", acl, "", const4acl.PermittedToEdit, false},
		{"nil ACL denies", ACL{}, "editor", const4acl.PermittedToEdit, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.acl.UserCan(tt.userID, tt.perm); got != tt.want {
				t.Errorf("UserCan(%q, %q) = %v, want %v", tt.userID, tt.perm, got, tt.want)
			}
		})
	}
}

func TestACL_Validate(t *testing.T) {
	tests := []struct {
		name    string
		acl     ACL
		wantErr bool
	}{
		{"empty", ACL{}, false},
		{"valid", ACL{Users: map[string]Permissions{"user1": {const4acl.PermittedToEdit: grant("user1")}}}, false},
		{"empty userID", ACL{Users: map[string]Permissions{"": {const4acl.PermittedToEdit: grant("user1")}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.acl.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
