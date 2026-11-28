package sneatcoretesting_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
	"github.com/sneat-co/sneat-go-core/sneatmemdb"
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
	if !errors.Is(err, sneatmemdb.ErrReadAfterWriteInTransaction) {
		t.Fatalf("NewInMemoryTestDB read-after-write error = %v, want %v (Firestore's transaction ordering)", err, sneatmemdb.ErrReadAfterWriteInTransaction)
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
	if !errors.Is(err, sneatmemdb.ErrReadAfterWriteInTransaction) {
		t.Fatalf("strict test DB read-after-write error = %v, want %v", err, sneatmemdb.ErrReadAfterWriteInTransaction)
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

// SneatProfile and WithSneatProfile are Firestore's bundle, not just any
// profile that orders reads before writes: a read-write transaction whose
// read is overtaken by another transaction's commit retries (optimistic
// concurrency), where a single-writer database would make the other wait.
// WithSneatProfile also resets a database built with another profile.
func TestSneatProfile_IsFirestore(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	base := ctx
	key := record.NewKeyWithID("c", "a")
	// attempts runs a transaction that reads key while another transaction
	// writes it, and returns how often the first one ran.
	attempts := func(t *testing.T, db dal.DB) int {
		t.Helper()
		if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			return tx.Set(ctx, record.NewRecordWithData(key, map[string]any{"v": 0}))
		}); err != nil {
			t.Fatal(err)
		}
		runs := 0
		err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			runs++
			read := map[string]any{}
			if err := tx.Get(ctx, record.NewRecordWithData(key, &read)); err != nil {
				return err
			}
			if runs == 1 {
				done := make(chan error, 1)
				go func() {
					// A separate transaction, not one nested in this callback.
					done <- db.RunReadwriteTransaction(base, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
						return tx.Set(ctx, record.NewRecordWithData(key, map[string]any{"v": 2}))
					})
				}()
				select {
				case err := <-done: // the other commit went through: ours must retry
					if err != nil {
						return err
					}
				case <-time.After(300 * time.Millisecond): // it waits for us: one writer
				}
			}
			return tx.Set(ctx, record.NewRecordWithData(key, map[string]any{"v": 1}))
		})
		if err != nil {
			t.Fatal(err)
		}
		return runs
	}
	if sneatcoretesting.SneatProfile() == nil {
		t.Fatal("SneatProfile() returned nil")
	}
	for name, db := range map[string]dal.DB{
		"WithSneatProfile":  sneatcoretesting.NewInMemoryTestDB(sneatcoretesting.WithSneatProfile()),
		"NewInMemoryTestDB": sneatcoretesting.NewInMemoryTestDB(),
	} {
		t.Run(name, func(t *testing.T) {
			if got := attempts(t, db); got != 2 {
				t.Errorf("an overtaken read-write transaction ran %d times, want 2 (Firestore retries it)", got)
			}
		})
	}
}
