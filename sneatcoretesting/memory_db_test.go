package sneatcoretesting_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dal-go/dalgo/adapters/dalgo2memory"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
)

func TestSetupMemoryDB(t *testing.T) {
	t.Parallel()
	ctx, db := sneatcoretesting.SetupMemoryDB(t)
	got, err := facade.GetSneatDB(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != db {
		t.Error("SetupMemoryDB did not install the created database")
	}
}

func TestSetupMemoryDB_ParallelIsolation(t *testing.T) {
	for _, name := range []string{"first", "second"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, db := sneatcoretesting.SetupMemoryDB(t)
			got, err := facade.GetSneatDB(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if got != db {
				t.Error("parallel test did not resolve its database")
			}
		})
	}
}

func TestNewInMemoryTestDB_RejectsReadsAfterWrites(t *testing.T) {
	t.Parallel()
	db := sneatcoretesting.NewInMemoryTestDB()
	key := record.NewKeyWithID("records", "firestore-ordering")
	err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		if err := tx.Set(ctx, record.NewRecordWithData(key, new(struct{}))); err != nil {
			return err
		}
		return tx.Get(ctx, record.NewRecordWithData(key, new(struct{})))
	})
	if !errors.Is(err, dalgo2memory.ErrReadAfterWriteInTransaction) {
		t.Fatalf("NewInMemoryTestDB read-after-write error = %v, want %v (Firestore's transaction ordering)", err, dalgo2memory.ErrReadAfterWriteInTransaction)
	}
}

func TestNewMemoryDB_RejectsReadsAfterWrites(t *testing.T) {
	t.Parallel()
	db := sneatcoretesting.NewMemoryDB()
	key := record.NewKeyWithID("records", "strict-ordering")
	err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		if err := tx.Set(ctx, record.NewRecordWithData(key, new(struct{}))); err != nil {
			return err
		}
		return tx.Get(ctx, record.NewRecordWithData(key, new(struct{})))
	})
	if !errors.Is(err, dalgo2memory.ErrReadAfterWriteInTransaction) {
		t.Fatalf("strict test DB read-after-write error = %v, want %v", err, dalgo2memory.ErrReadAfterWriteInTransaction)
	}
}

func TestNewStrictSchemaMemoryDB_RejectsUndefinedCollections(t *testing.T) {
	t.Parallel()
	db := sneatcoretesting.NewStrictSchemaMemoryDB()
	err := db.Get(context.Background(), record.NewRecordWithData(
		record.NewKeyWithID("undefined", "record"),
		new(struct{}),
	))
	if err == nil {
		t.Fatal("strict schema test DB accepted an undefined collection")
	}
}
