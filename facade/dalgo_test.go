package facade

import (
	"context"
	"errors"
	"github.com/dal-go/dalgo/dal"
	"testing"
)

func TestGetDatabase(t *testing.T) {
	db, err := GetDatabase(context.Background())
	if db != nil {
		t.Error("GetDatabase() = ", db)
	}
	if !errors.Is(err, ErrNotInitialized) {
		t.Error("Expected to return ErrNotInitialized")
	}
}

func TestRunReadwriteTransaction(t *testing.T) {
	ctx := context.Background()
	err := RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return nil
	})
	if !errors.Is(err, ErrNotInitialized) {
		t.Error("Expected to return ErrNotInitialized")
	}
}
