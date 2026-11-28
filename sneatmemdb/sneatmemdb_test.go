package sneatmemdb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/dal-go/dalgo/adapters/dalgo2memory"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/sneat-co/sneat-go-core/sneatmemdb"
)

func TestNew_FirestoreReadAfterWriteRejected(t *testing.T) {
	t.Parallel()
	db := sneatmemdb.New()
	key := record.NewKeyWithID("records", "firestore-ordering")
	err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		if err := tx.Set(ctx, record.NewRecordWithData(key, new(struct{}))); err != nil {
			return err
		}
		return tx.Get(ctx, record.NewRecordWithData(key, new(struct{})))
	})
	if !errors.Is(err, sneatmemdb.ErrReadAfterWriteInTransaction) {
		t.Fatalf("New() read-after-write error = %v, want %v", err, sneatmemdb.ErrReadAfterWriteInTransaction)
	}
}

func TestNew_TimePrecisionIsMicroseconds(t *testing.T) {
	t.Parallel()
	type stamped struct {
		At time.Time
	}
	db := sneatmemdb.New()
	ctx := context.Background()
	key := record.NewKeyWithID("timestamps", "micro")
	want := time.Date(2026, 11, 28, 9, 30, 45, 123456789, time.UTC)

	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Set(ctx, record.NewRecordWithData(key, &stamped{At: want}))
	}); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	var got stamped
	if err := db.Get(ctx, record.NewRecordWithData(key, &got)); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	wantTruncated := want.Truncate(time.Microsecond)
	if !got.At.Equal(wantTruncated) {
		t.Fatalf("stored time = %v, want %v", got.At, wantTruncated)
	}
	if got.At.Equal(want) {
		t.Fatalf("stored time preserved nanoseconds (%v), expected microsecond truncation", got.At)
	}
}

func TestWithSchemaAndWithCollection(t *testing.T) {
	t.Parallel()
	type item struct {
		Name string
	}
	db := sneatmemdb.New(
		sneatmemdb.WithSchema(
			false,
			sneatmemdb.WithCollection("items", func() *item { return &item{} }),
		),
	)
	ctx := context.Background()

	declaredKey := record.NewKeyWithID("items", "item-1")
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Set(ctx, record.NewRecordWithData(declaredKey, &item{Name: "declared"}))
	}); err != nil {
		t.Fatalf("Set on declared collection failed: %v", err)
	}
	var got item
	if err := db.Get(ctx, record.NewRecordWithData(declaredKey, &got)); err != nil {
		t.Fatalf("Get on declared collection failed: %v", err)
	}
	if got.Name != "declared" {
		t.Fatalf("got Name = %q, want %q", got.Name, "declared")
	}

	undeclaredKey := record.NewKeyWithID("undeclared", "record-1")
	if err := db.Get(ctx, record.NewRecordWithData(undeclaredKey, new(struct{}))); err == nil {
		t.Fatal("strict schema accepted an undeclared collection")
	}
}

// SneatProfile and WithSneatProfile are Firestore's bundle, not just any
// profile that orders reads before writes: a read-write transaction whose
// read is overtaken by another transaction's commit retries (optimistic
// concurrency), where a single-writer database would make the other wait.
// WithSneatProfile also resets a database built with another profile.
func TestSneatProfileAndWithSneatProfile_IsFirestore(t *testing.T) {
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
	if got := attempts(t, dalgo2memory.New(dalgo2memory.SingleWriterProfile())); got != 1 {
		t.Fatalf("control: a single-writer transaction ran %d times, want 1", got)
	}
	for name, db := range map[string]dal.DB{
		"SneatProfile":     dalgo2memory.New(sneatmemdb.SneatProfile()),
		"WithSneatProfile": dalgo2memory.New(dalgo2memory.SingleWriterProfile(), sneatmemdb.WithSneatProfile()),
		"New":              sneatmemdb.New(),
	} {
		t.Run(name, func(t *testing.T) {
			if got := attempts(t, db); got != 2 {
				t.Errorf("an overtaken read-write transaction ran %d times, want 2 (Firestore retries it)", got)
			}
		})
	}
}

func TestNewBranchingProvider_Capture(t *testing.T) {
	t.Parallel()
	provider := sneatmemdb.NewBranchingProvider()
	if provider == nil {
		t.Fatal("NewBranchingProvider() returned nil")
	}
	db := sneatmemdb.New()
	ctx := context.Background()
	key := record.NewKeyWithID("items", "branch-1")
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Set(ctx, record.NewRecordWithData(key, map[string]any{"v": 1}))
	}); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	checkpoint, err := provider.Capture(ctx, db)
	if err != nil {
		t.Fatalf("Capture failed: %v", err)
	}
	if checkpoint == nil {
		t.Fatal("Capture returned nil checkpoint")
	}
}

func TestErrorVarsAndIsTransactionConflict(t *testing.T) {
	t.Parallel()
	if sneatmemdb.ErrReadAfterWriteInTransaction != dalgo2memory.ErrReadAfterWriteInTransaction {
		t.Fatal("ErrReadAfterWriteInTransaction does not match dalgo2memory.ErrReadAfterWriteInTransaction")
	}
	if sneatmemdb.ErrTransactionConflict != dalgo2memory.ErrTransactionConflict {
		t.Fatal("ErrTransactionConflict does not match dalgo2memory.ErrTransactionConflict")
	}
	if sneatmemdb.ErrNestedTransaction != dalgo2memory.ErrNestedTransaction {
		t.Fatal("ErrNestedTransaction does not match dalgo2memory.ErrNestedTransaction")
	}
	if !sneatmemdb.IsTransactionConflict(sneatmemdb.ErrTransactionConflict) {
		t.Fatal("IsTransactionConflict(ErrTransactionConflict) = false, want true")
	}
	if !sneatmemdb.IsTransactionConflict(fmt.Errorf("wrapped: %w", sneatmemdb.ErrTransactionConflict)) {
		t.Fatal("IsTransactionConflict(wrapped ErrTransactionConflict) = false, want true")
	}
	if sneatmemdb.IsTransactionConflict(errors.New("unrelated")) {
		t.Fatal("IsTransactionConflict(unrelated) = true, want false")
	}
}
