package facade

import (
	"context"
	"errors"
	"github.com/dal-go/dalgo/dal"
	"testing"
)

func TestGetDatabase(t *testing.T) {
	ctx := context.Background()
	defer mustPanicWithErrNotInitialized(t)
	_, _ = GetDatabase(ctx)
	//if db != nil {
	//	t.Error("GetDatabase() = ", db)
	//}
	//if !errors.Is(err, ErrNotInitialized) {
	//	t.Error("Expected to return ErrNotInitialized")
	//}
}

func TestRunReadwriteTransaction(t *testing.T) {
	defer mustPanicWithErrNotInitialized(t)
	ctx := context.Background()
	_ = RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return nil
	})
}

func mustPanicWithErrNotInitialized(t *testing.T) {
	if r := recover(); r == nil {
		t.Error("Expected to panic")
	} else if err := r.(error); !errors.Is(err, ErrNotInitialized) {
		t.Error("Expected to return ErrNotInitialized")
	}
}
