// Copyright 2026 Sneat Co.
package firestorepolicy

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/dal-go/dalgo/access"
	"github.com/dal-go/dalgo/adapters/dalgo2memory"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
)

// Queued email records live at ext/emailing/emails/{id}, owned by the
// sneat-core-modules emailing package. The former root EmailAddress stays
// unapproved.
func TestEmailingExtensionPlacement(t *testing.T) {
	raw := dalgo2memory.New(dalgo2memory.FirestoreProfile())
	db, err := ProtectRootCollections(raw)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	parent := record.NewKeyWithID("ext", "emailing")

	// Matches sneat-core-modules models4auth.NewEmailKey("") with an adapter-generated ID.
	generated := record.NewRecordWithData(record.NewIncompleteKey("emails", reflect.String, parent), map[string]any{"Status": "queued"})
	if err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Insert(ctx, generated, dal.WithAdapterGeneratedID())
	}); err != nil {
		t.Fatalf("registered emailing insert refused: %v", err)
	}
	if id, ok := generated.Key().ID.(string); !ok || id == "" {
		t.Fatalf("generated email ID missing: %v", generated.Key())
	}
	if exists, err := raw.Exists(ctx, generated.Key()); err != nil || !exists {
		t.Fatalf("approved email row not stored %s: %v", generated.Key(), err)
	}
	if err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Set(ctx, record.NewRecordWithData(generated.Key(), map[string]any{"Status": "sent"}))
	}); err != nil {
		t.Fatalf("registered emailing set refused: %v", err)
	}

	for _, key := range []*record.Key{
		record.NewIncompleteKey("EmailAddress", reflect.String, nil),
		record.NewKeyWithID("EmailAddress", "email1"),
		record.NewKeyWithID("emails", "email1"),
		record.NewKeyWithParentAndID(record.NewKeyWithID("ext", "unknown-emailing"), "emails", "email1"),
	} {
		row := record.NewRecordWithData(key, map[string]any{"Status": "queued"})
		err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			if key.ID == nil {
				return tx.Insert(ctx, row, dal.WithAdapterGeneratedID())
			}
			return tx.Set(ctx, row)
		})
		if !errors.Is(err, access.ErrAccessDenied) {
			t.Fatalf("unregistered email placement accepted %s: %v", key, err)
		}
	}
}
