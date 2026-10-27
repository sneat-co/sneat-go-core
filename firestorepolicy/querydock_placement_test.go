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
	"github.com/dal-go/record/update"
)

// Placement is mandatory even when a request-bound policy permits all writes;
// extension registration does not confer caller ownership or client access.
func TestQueryDockMandatoryPlacementAndCallerIsolation(t *testing.T) {
	raw := dalgo2memory.New(dalgo2memory.FirestoreProfile())
	db, err := ProtectRootCollections(raw)
	if err != nil {
		t.Fatal(err)
	}
	ctx := access.WithPolicy(context.Background(), access.MustPolicy("caller-allows", access.Root(access.Allow(access.ReadWrite))))
	extension := record.NewKeyWithID("ext", "querydock")
	fiddle := record.NewKeyWithParentAndID(extension, "fiddles", "fiddle-1")
	for _, key := range []*record.Key{
		fiddle,
		record.NewKeyWithParentAndID(fiddle, "revisions", "1"),
		record.NewKeyWithParentAndID(extension, "usage", "user-day"),
	} {
		row := record.NewRecordWithData(key, map[string]any{"n": 1})
		if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.Insert(ctx, row) }); err != nil {
			t.Fatalf("approved insert %s: %v", key, err)
		}
		if exists, err := raw.Exists(ctx, key); err != nil || !exists {
			t.Fatalf("approved row not stored %s: %v", key, err)
		}
		if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			return tx.Update(ctx, key, []update.Update{update.ByFieldName("n", 2)})
		}); err != nil {
			t.Fatalf("approved update %s: %v", key, err)
		}
		if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.Delete(ctx, key) }); err != nil {
			t.Fatalf("approved delete %s: %v", key, err)
		}
		if exists, err := raw.Exists(ctx, key); err != nil || exists {
			t.Fatalf("approved delete did not persist %s: %v", key, err)
		}
	}
	unknown := record.NewKeyWithParentAndID(record.NewKeyWithID("ext", "unknown-querydock-extension"), "fiddles", "fiddle-1")
	for _, key := range []*record.Key{unknown, record.NewKeyWithParentAndID(unknown, "revisions", "1"), record.NewKeyWithParentAndID(unknown.Parent(), "usage", "user-day"), record.NewKeyWithID("querydock", "forbidden-root")} {
		row := record.NewRecordWithData(key, map[string]any{"n": 1})
		if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.Set(ctx, row) }); !errors.Is(err, access.ErrAccessDenied) {
			t.Fatalf("unregistered placement accepted %s: %v", key, err)
		}
		if exists, err := raw.Exists(ctx, key); err != nil || exists {
			t.Fatalf("denied row reached storage %s: %v", key, err)
		}
	}
	good := record.NewRecordWithData(fiddle, map[string]any{"n": 1})
	bad := record.NewRecordWithData(unknown, map[string]any{"n": 1})
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.SetMulti(ctx, []record.Record{good, bad})
	}); !errors.Is(err, access.ErrAccessDenied) {
		t.Fatalf("mixed batch accepted: %v", err)
	}
	if exists, _ := raw.Exists(ctx, fiddle); exists {
		t.Fatal("mixed batch partially persisted approved QueryDock row")
	}
	denied := access.WithPolicy(context.Background(), access.MustPolicy("caller-denies", access.Root(access.Deny(access.Write))))
	if err := db.RunReadwriteTransaction(denied, func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.Set(ctx, good) }); !errors.Is(err, access.ErrAccessDenied) {
		t.Fatalf("placement widened caller authorization: %v", err)
	}
	if exists, _ := raw.Exists(ctx, fiddle); exists {
		t.Fatal("caller-denied QueryDock row reached storage")
	}
}
