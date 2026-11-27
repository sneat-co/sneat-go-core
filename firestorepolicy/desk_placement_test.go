// Copyright 2026 Sneat Co.
package firestorepolicy

import (
	"context"
	"errors"
	"testing"

	"github.com/dal-go/dalgo/access"
	"github.com/dal-go/dalgo/adapters/dalgo2memory"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
)

// Sneat Dev Desk (github.com/sneat-dev/desk/backend) keeps its device grants
// and machine credentials under ext/desk, and machines, snapshots and audit
// events under spaces/{space}/ext/desk. Its former top-level collection
// names stay unapproved.
func TestDeskExtensionPlacement(t *testing.T) {
	raw := dalgo2memory.New(dalgo2memory.FirestoreProfile())
	db, err := ProtectRootCollections(raw)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	global := record.NewKeyWithID("ext", "desk")
	space := record.NewKeyWithParentAndID(record.NewKeyWithID("spaces", "space1"), "ext", "desk")
	machine := record.NewKeyWithParentAndID(space, "machines", "owner1-machine1")
	for _, key := range []*record.Key{
		record.NewKeyWithParentAndID(global, "desk_grants", "grant1"),
		record.NewKeyWithParentAndID(global, "desk_credentials", "credential1"),
		machine,
		record.NewKeyWithParentAndID(machine, "desk_snapshot_history", "snapshot1"),
	} {
		row := record.NewRecordWithData(key, map[string]any{"v": 1})
		if err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			return tx.Set(ctx, row)
		}); err != nil {
			t.Fatalf("registered desk placement refused %s: %v", key, err)
		}
	}
	for _, key := range []*record.Key{
		record.NewKeyWithID("desk_grants", "grant1"),
		record.NewKeyWithID("desk_credentials", "credential1"),
	} {
		row := record.NewRecordWithData(key, map[string]any{"v": 1})
		err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			return tx.Set(ctx, row)
		})
		if !errors.Is(err, access.ErrAccessDenied) {
			t.Fatalf("top-level desk placement accepted %s: %v", key, err)
		}
	}
}
