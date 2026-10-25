// Copyright 2026 Sneat Co.
package facade

import (
	"context"
	"errors"
	"testing"

	"github.com/dal-go/dalgo/access"
	"github.com/dal-go/dalgo/adapters/dalgo2memory"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
)

type firestoreNamedDB struct{ dal.DB }

func (firestoreNamedDB) Adapter() dal.Adapter { return dal.NewAdapter("firestore", "test") }

func TestProtectSneatDBAndContextOverride(t *testing.T) {
	ctx := context.Background()
	raw := dalgo2memory.New(dalgo2memory.FirestoreProfile())
	if _, err := ProtectSneatDB(nil); err == nil {
		t.Fatal("nil accepted")
	}
	if got, err := ProtectSneatDB(raw); err != nil || got != raw {
		t.Fatal("non-Firestore changed", err)
	}
	firestoreDB := firestoreNamedDB{raw}
	protected, err := ProtectSneatDB(firestoreDB)
	if err != nil {
		t.Fatal(err)
	}
	denied := func(db dal.DB) {
		t.Helper()
		err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			return tx.Set(ctx, record.NewRecordWithData(record.NewKeyWithID("unapprovedRoot", "bad"), map[string]any{"n": 1}))
		})
		if !errors.Is(err, access.ErrAccessDenied) {
			t.Fatalf("unguarded handle: %v", err)
		}
	}
	denied(protected)
	got, err := GetSneatDB(WithSneatDB(ctx, firestoreDB))
	if err != nil {
		t.Fatal(err)
	}
	denied(got)
	defaultSneatDBProviderMu.RLock()
	old := defaultSneatDBProvider
	defaultSneatDBProviderMu.RUnlock()
	t.Cleanup(func() { SetDefaultSneatDBProvider(old) })
	SetDefaultSneatDBProvider(func(context.Context) (dal.DB, error) { return firestoreDB, nil })
	got, err = GetSneatDB(ctx)
	if err != nil {
		t.Fatal(err)
	}
	denied(got)
	if _, err = GetSneatDB(WithSneatDBProvider(ctx, func(context.Context) (dal.DB, error) { return nil, nil })); err == nil {
		t.Fatal("nil provider result accepted")
	}
	oldProtect := protectRootCollections
	protectRootCollections = func(dal.DB) (dal.DB, error) { return nil, errors.New("registry unavailable") }
	t.Cleanup(func() { protectRootCollections = oldProtect })
	if _, err = ProtectSneatDB(firestoreDB); err == nil {
		t.Fatal("policy failure bypassed")
	}
}
