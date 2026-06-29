package dbo4sharing

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/sneat-co/sneat-go-core/acl/const4acl"
	"github.com/sneat-co/sneat-go-core/acl/dbo4acl"
	"github.com/strongo/strongoapp/with"
)

// TestTo_EmbeddedACLIsPromoted guards the firestore wire format. Firestore (like
// encoding/json) promotes the fields of an anonymously-embedded struct to the
// parent ONLY when that embedded field has no struct tag. If someone adds a
// `firestore:"acl"` (or json) tag to the embedded ACL, `users` would silently
// nest under `acl`/`ACL` and every existing happening `sharedTo` document would
// stop loading. This asserts the promotion precondition directly, covering the
// firestore code path that the JSON marshal test below cannot reach.
func TestTo_EmbeddedACLIsPromoted(t *testing.T) {
	f, ok := reflect.TypeOf(To{}).FieldByName("ACL")
	if !ok {
		t.Fatal("To no longer embeds dbo4acl.ACL — wire format at risk")
	}
	if !f.Anonymous {
		t.Error("ACL must be embedded anonymously so its `users` field is promoted")
	}
	if tag, has := f.Tag.Lookup("firestore"); has {
		t.Errorf("embedded ACL must NOT carry a firestore tag (got %q) — it would nest `users`", tag)
	}
	if tag, has := f.Tag.Lookup("json"); has {
		t.Errorf("embedded ACL must NOT carry a json tag (got %q) — it would nest `users`", tag)
	}
}

// TestTo_JSONWireFormat guards the wire format: factoring the per-user grants
// into an embedded dbo4acl.ACL must NOT change the JSON shape, so existing
// happening `sharedTo` documents keep round-tripping. The grants must stay at
// `users`, never nested under an `ACL`/`acl` key.
func TestTo_JSONWireFormat(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := To{
		ACL: dbo4acl.ACL{
			Users: map[string]dbo4acl.Permissions{
				"user1": {
					const4acl.PermittedToEdit: with.CreatedFields{
						CreatedAtField: with.CreatedAtField{CreatedAt: now},
						CreatedByField: with.CreatedByField{CreatedBy: "user1"},
					},
				},
			},
		},
	}
	data, err := json.Marshal(to)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, `"users"`) {
		t.Errorf("expected top-level \"users\" key, got: %s", got)
	}
	if strings.Contains(got, `"ACL"`) || strings.Contains(got, `"acl"`) {
		t.Errorf("grants leaked under an ACL key (wire format changed): %s", got)
	}

	// And it must unmarshal back through the promoted Users field.
	var back To
	if err = json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !back.UserCan("user1", const4acl.PermittedToEdit) {
		t.Errorf("round-trip lost the user grant: %s", got)
	}
}
