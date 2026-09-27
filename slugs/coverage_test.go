package slugs_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/dal-go/record/update"
	"github.com/sneat-co/sneat-go-core/slugs"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
)

type errorGetter struct {
	dal.Getter
	err error
}

func (g errorGetter) Get(_ context.Context, _ record.Record) error {
	return g.err
}

func TestResolve_StorageError(t *testing.T) {
	expectedErr := errors.New("storage get failed")
	_, err := slugs.Resolve(context.Background(), errorGetter{err: expectedErr}, "ns", "valid-slug")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

type errorUpdater struct {
	dal.ReadwriteTransaction
	err error
}

func (u errorUpdater) Update(_ context.Context, _ *record.Key, _ []update.Update, _ ...dal.Precondition) error {
	return u.err
}

func TestRelease_StorageError(t *testing.T) {
	expectedErr := errors.New("storage update failed")
	err := slugs.Release(context.Background(), errorUpdater{err: expectedErr}, "ns", "valid-slug")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

type renameErrorTx struct {
	dal.ReadwriteTransaction
	failOnUpdate bool
}

func (r *renameErrorTx) Update(ctx context.Context, key *record.Key, updates []update.Update, preconditions ...dal.Precondition) error {
	if r.failOnUpdate {
		return errors.New("update tombstone failed")
	}
	return r.ReadwriteTransaction.Update(ctx, key, updates, preconditions...)
}

func TestRename_TombstoneError(t *testing.T) {
	db := sneatcoretesting.NewMemoryDB()
	ctx := context.Background()

	_ = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		txWrapper := &renameErrorTx{ReadwriteTransaction: tx, failOnUpdate: true}
		_, err := slugs.Rename(ctx, txWrapper, "ns", "old-slug", "new-slug", "target1", "kind")
		if err == nil {
			t.Fatal("expected error on tombstone failure, got nil")
		}
		return nil
	})
}

type errorExecutor struct {
	dal.QueryExecutor
	err error
}

func (e errorExecutor) ExecuteQueryToRecordsReader(_ context.Context, _ dal.Query) (dal.RecordsReader, error) {
	return nil, e.err
}

type errorReaderExecutor struct {
	dal.QueryExecutor
}

func (errorReaderExecutor) ExecuteQueryToRecordsReader(_ context.Context, _ dal.Query) (dal.RecordsReader, error) {
	return errorRecordsReader{}, nil
}

type errorRecordsReader struct {
	dal.RecordsReader
}

func (errorRecordsReader) Next() (record.Record, error) {
	return nil, errors.New("reader stream failed")
}

func (errorRecordsReader) Close() error {
	return nil
}

func TestEnumerate_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("query_error", func(t *testing.T) {
		_, err := slugs.Enumerate(ctx, errorExecutor{err: errors.New("query failed")}, "ns")
		if err == nil {
			t.Fatal("expected query error, got nil")
		}
	})

	t.Run("reader_error", func(t *testing.T) {
		_, err := slugs.Enumerate(ctx, errorReaderExecutor{}, "ns")
		if err == nil {
			t.Fatal("expected reader error, got nil")
		}
	})
}
